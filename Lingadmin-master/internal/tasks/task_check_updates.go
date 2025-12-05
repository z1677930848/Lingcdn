// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package tasks

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	teaconst "github.com/TeaOSLab/EdgeAdmin/internal/const"
	"github.com/TeaOSLab/EdgeAdmin/internal/events"
	"github.com/TeaOSLab/EdgeAdmin/internal/goman"
	"github.com/TeaOSLab/EdgeAdmin/internal/rpc"
	"github.com/TeaOSLab/EdgeAdmin/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	stringutil "github.com/iwind/TeaGo/utils/string"
)

var (
	metaHTTPClient     = &http.Client{Timeout: 30 * time.Second}
	downloadHTTPClient = &http.Client{Timeout: 30 * time.Minute}
)

type UpdateInfo struct {
	Version        string `json:"version"`
	CurrentVersion string `json:"currentVersion"`
	DownloadURL    string `json:"downloadURL"`
	Changelog      string `json:"changelog"`
	Description    string `json:"description"`
	SHA256         string `json:"sha256"`
	CheckTime      string `json:"checkTime"`
}

func init() {
	events.On(events.EventStart, func() {
		var task = NewCheckUpdatesTask()
		goman.New(func() {
			task.Start()
		})

		// 閸氼垰濮╂稉瀛樻閺傚洣娆㈠〒鍛倞娴犺濮?
		utils.ScheduleCleanupTask()
	})
}

type CheckUpdatesTask struct {
	ticker     *time.Ticker
	logManager *utils.UpgradeLogManager
	notifier   utils.UpdateNotifier
	cleaner    *utils.TempFileCleaner
}

func NewCheckUpdatesTask() *CheckUpdatesTask {
	// 閸掓稑缂撴径姘垛偓姘朵壕闁氨鐓￠崳?
	multiNotifier := utils.NewMultiNotifier()
	multiNotifier.AddNotifier(utils.NewLogNotifier())
	// 閸欘垱鐗撮幑顕€鍘ょ純顔藉潑閸旂姵娲挎径姘垛偓姘辩叀閸?
	// multiNotifier.AddNotifier(utils.NewWebhookNotifier("http://your-webhook-url"))

	return &CheckUpdatesTask{
		logManager: utils.SharedUpgradeLogManager(),
		notifier:   multiNotifier,
		cleaner:    utils.NewTempFileCleaner(),
	}
}

func (this *CheckUpdatesTask) Start() {
	// 閸氼垰濮╅崥搴ｇ彌閸楄櫕顥呴弻銉ょ濞?
	err := this.Loop()
	if err != nil {
		logs.Println("[TASK][CHECK_UPDATES_TASK]" + err.Error())
	}

	// 閻掕泛鎮楀В?鐏忓繑妞傚Λ鈧弻銉ょ濞?
	this.ticker = time.NewTicker(6 * time.Hour)
	for range this.ticker.C {
		err := this.Loop()
		if err != nil {
			logs.Println("[TASK][CHECK_UPDATES_TASK]" + err.Error())
		}
	}
}

func (this *CheckUpdatesTask) Loop() error {
	// 濡偓閺屻儲妲搁崥锕€绱戦崥?
	rpcClient, err := rpc.SharedRPC()
	if err != nil {
		return err
	}
	valueResp, err := rpcClient.SysSettingRPC().ReadSysSetting(rpcClient.Context(0), &pb.ReadSysSettingRequest{Code: systemconfigs.SettingCodeCheckUpdates})
	if err != nil {
		return err
	}
	var valueJSON = valueResp.ValueJSON
	var config = &systemconfigs.CheckUpdatesConfig{AutoCheck: false}
	if len(valueJSON) > 0 {
		err = json.Unmarshal(valueJSON, config)
		if err != nil {
			return errors.New("decode config failed: " + err.Error())
		}
		if !config.AutoCheck {
			return nil
		}
	} else {
		return nil
	}

	// 瀵偓婵顥呴弻?
	type Response struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}

	// 閻╊喖澧犻弨顖涘瘮Linux
	if runtime.GOOS != "linux" {
		return nil
	}

	var apiURL = teaconst.UpdatesURL
	apiURL = strings.ReplaceAll(apiURL, "${os}", runtime.GOOS)
	apiURL = strings.ReplaceAll(apiURL, "${arch}", runtime.GOARCH)

	logs.Println("[TASK][CHECK_UPDATES_TASK]checking updates from:", apiURL)

	resp, err := metaHTTPClient.Get(apiURL)
	if err != nil {
		return errors.New("read api failed: " + err.Error())
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("read api failed: http status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("read api failed: " + err.Error())
	}

	var apiResponse = &Response{}
	err = json.Unmarshal(data, apiResponse)
	if err != nil {
		return errors.New("decode version data failed: " + err.Error())
	}

	if apiResponse.Code != 200 {
		return errors.New("invalid response: " + apiResponse.Message)
	}

	var m = maps.NewMap(apiResponse.Data)
	var dlHost = m.GetString("host")
	var versions = m.GetSlice("versions")
	if len(versions) > 0 {
		for _, version := range versions {
			var vMap = maps.NewMap(version)
			if vMap.GetString("code") == "admin" {
				var latestVersion = vMap.GetString("version")
				var changelog = vMap.GetString("changelog")
				var description = vMap.GetString("description")
				var downloadURL = dlHost + vMap.GetString("url")
				var fileSHA256 = vMap.GetString("sha256")

				logs.Println("[TASK][CHECK_UPDATES_TASK]current version:", teaconst.Version, "latest version:", latestVersion)

				if stringutil.VersionCompare(teaconst.Version, latestVersion) < 0 {
					teaconst.NewVersionCode = latestVersion
					teaconst.NewVersionDownloadURL = downloadURL

					// 娣囨繂鐡ㄩ弴瀛樻煀娣団剝浼呴崚鐗堟瀮娴?
					updateInfo := &UpdateInfo{
						Version:        latestVersion,
						CurrentVersion: teaconst.Version,
						DownloadURL:    downloadURL,
						Changelog:      changelog,
						Description:    description,
						SHA256:         fileSHA256,
						CheckTime:      time.Now().Format("2006-01-02 15:04:05"),
					}
					updateInfoJSON, _ := json.MarshalIndent(updateInfo, "", "  ")
					_ = os.WriteFile(Tea.ConfigFile("update_info.json"), updateInfoJSON, 0644)

					logs.Println("[TASK][CHECK_UPDATES_TASK]new version available:", latestVersion)
					logs.Println("[TASK][CHECK_UPDATES_TASK]download url:", downloadURL)
					logs.Println("[TASK][CHECK_UPDATES_TASK]changelog:", changelog)

					return nil
				} else {
					logs.Println("[TASK][CHECK_UPDATES_TASK]no updates available, current version is latest")
					teaconst.NewVersionCode = ""
					teaconst.NewVersionDownloadURL = ""
				}
			}
		}
	}

	return nil
}

// DownloadAndInstallUpdate 娑撳娴囬獮璺虹暔鐟佸懏娲块弬甯礄閺€纭呯箻閻楀牞绱?
func DownloadAndInstallUpdate() error {
	startTime := time.Now()
	logs.Println("[UPDATE]starting update process...")

	// 閸掓稑缂撻崡鍥╅獓閺冦儱绻?
	logManager := utils.SharedUpgradeLogManager()
	upgradeLog := &utils.UpgradeLog{
		Component:  "admin",
		OldVersion: teaconst.Version,
		Status:     utils.StatusPending,
		StartTime:  startTime,
	}
	_ = logManager.CreateLog(upgradeLog)

	// 閸掓稑缂撴稉瀛樻閺傚洣娆㈠〒鍛倞閸?
	cleaner := utils.NewTempFileCleaner()
	defer func() {
		if err := cleaner.CleanupAll(); err != nil {
			logs.Println("[UPDATE]cleanup failed:", err)
		}
	}()

	// 閸掓稑缂撻柅姘辩叀閸?
	notifier := utils.NewMultiNotifier()
	notifier.AddNotifier(utils.NewLogNotifier())
	notifier.AddNotifier(utils.NewConsoleNotifier())

	// ??????
	updateInfoData, err := os.ReadFile(Tea.ConfigFile("update_info.json"))
	if err != nil {
		upgradeErr := utils.NewUpgradeError(utils.StageCheckVersion, utils.ErrCodeNetworkFailed,
			"read update info failed", err)
		handleUpgradeError(upgradeLog, logManager, notifier, upgradeErr)
		return upgradeErr
	}

	var updateInfo UpdateInfo
	err = json.Unmarshal(updateInfoData, &updateInfo)
	if err != nil {
		upgradeErr := utils.NewUpgradeError(utils.StageCheckVersion, utils.ErrCodeInvalidResponse,
			"parse update info failed", err)
		handleUpgradeError(upgradeLog, logManager, notifier, upgradeErr)
		return upgradeErr
	}

	if updateInfo.DownloadURL == "" || updateInfo.Version == "" {
		upgradeErr := utils.NewUpgradeError(utils.StageCheckVersion, utils.ErrCodeInvalidResponse,
			"download url or version missing in update info", nil)
		handleUpgradeError(upgradeLog, logManager, notifier, upgradeErr)
		return upgradeErr
	}

	if updateInfo.SHA256 == "" {
		upgradeErr := utils.NewUpgradeError(utils.StageVerify, utils.ErrCodeVerifyFailed,
			"sha256 not provided in update info", nil)
		handleUpgradeError(upgradeLog, logManager, notifier, upgradeErr)
		return upgradeErr
	}

	downloadURL := updateInfo.DownloadURL
	expectedSHA256 := updateInfo.SHA256
	version := updateInfo.Version
	upgradeLog.NewVersion = version
	upgradeLog.DownloadURL = downloadURL
	_ = logManager.UpdateLog(upgradeLog)

	// 闁氨鐓″鈧慨?
	notifier.NotifyStart("admin", version)

	logs.Println("[UPDATE]downloading version:", version)
	logs.Println("[UPDATE]download url:", downloadURL)

	// 閸掓稑缂撴稉瀛樻閻╊喖缍?
	tmpDir := Tea.ConfigFile("tmp")
	_ = os.MkdirAll(tmpDir, 0755)

	// 娑撳娴囬弬鍥︽
	downloadFilePath := filepath.Join(tmpDir, fmt.Sprintf("ling-admin-v%s.zip", version))
	cleaner.AddFile(downloadFilePath)

	upgradeLog.Status = utils.StatusDownloading
	_ = logManager.UpdateLog(upgradeLog)

	err = downloadFileWithProgress(downloadHTTPClient, downloadURL, downloadFilePath, func(progress float32, speed float64) {
		message := fmt.Sprintf("Downloading: %.1f MB/s", speed)
		normalizedProgress := progress
		if normalizedProgress < 0 {
			normalizedProgress = 0
		}
		notifier.NotifyProgress("admin", normalizedProgress*0.6, message)
	})
	if err != nil {
		upgradeErr := utils.NewUpgradeError(utils.StageDownload, utils.ErrCodeDownloadFailed,
			"download failed", err)
		handleUpgradeError(upgradeLog, logManager, notifier, upgradeErr)
		return upgradeErr
	}

	stat, _ := os.Stat(downloadFilePath)
	upgradeLog.DownloadSize = stat.Size()
	_ = logManager.UpdateLog(upgradeLog)

	logs.Println("[UPDATE]download completed")
	notifier.NotifyProgress("admin", 0.65, "Download complete, verifying...")

	// 妤犲矁鐦塖HA256
	upgradeLog.Status = utils.StatusVerifying
	_ = logManager.UpdateLog(upgradeLog)

	if expectedSHA256 != "" {
		actualSHA256, err := calculateSHA256(downloadFilePath)
		if err != nil {
			upgradeErr := utils.NewUpgradeError(utils.StageVerify, utils.ErrCodeVerifyFailed,
				"calculate sha256 failed", err)
			handleUpgradeError(upgradeLog, logManager, notifier, upgradeErr)
			return upgradeErr
		}
		if actualSHA256 != expectedSHA256 {
			upgradeErr := utils.NewUpgradeError(utils.StageVerify, utils.ErrCodeVerifyFailed,
				fmt.Sprintf("sha256 mismatch: expected %s, got %s", expectedSHA256, actualSHA256), nil)
			handleUpgradeError(upgradeLog, logManager, notifier, upgradeErr)
			return upgradeErr
		}
		logs.Println("[UPDATE]sha256 verification passed")
		notifier.NotifyProgress("admin", 0.7, "Verification passed")
	}

	// 鐟欙絽甯囬弬鍥︽
	extractDir := filepath.Join(tmpDir, "extract")
	_ = os.RemoveAll(extractDir)
	_ = os.MkdirAll(extractDir, 0755)
	cleaner.AddDir(extractDir)

	notifier.NotifyProgress("admin", 0.75, "Extracting files...")
	err = unzip(downloadFilePath, extractDir)
	if err != nil {
		upgradeErr := utils.NewUpgradeError(utils.StageUnzip, utils.ErrCodeUnzipFailed,
			"unzip failed", err)
		handleUpgradeError(upgradeLog, logManager, notifier, upgradeErr)
		return upgradeErr
	}

	logs.Println("[UPDATE]extract completed")
	notifier.NotifyProgress("admin", 0.85, "Installing...")

	// 閹垫儳鍩屾禍宀冪箻閸掕埖鏋冩禒?
	binaryPath := filepath.Join(extractDir, "ling-admin")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		upgradeErr := utils.NewUpgradeError(utils.StageInstall, utils.ErrCodeInstallFailed,
			"binary file not found in package", nil)
		handleUpgradeError(upgradeLog, logManager, notifier, upgradeErr)
		return upgradeErr
	}

	upgradeLog.Status = utils.StatusInstalling
	_ = logManager.UpdateLog(upgradeLog)

	// 婢跺洣鍞よぐ鎾冲閻楀牊婀?
	currentBinary, err := os.Executable()
	if err != nil {
		upgradeErr := utils.NewUpgradeError(utils.StageBackup, utils.ErrCodeBackupFailed,
			"get current binary path failed", err)
		handleUpgradeError(upgradeLog, logManager, notifier, upgradeErr)
		return upgradeErr
	}

	backupPath := currentBinary + ".backup." + teaconst.Version
	if err := copyFile(currentBinary, backupPath); err != nil {
		upgradeErr := utils.NewUpgradeError(utils.StageBackup, utils.ErrCodeBackupFailed,
			"backup current binary failed", err)
		handleUpgradeError(upgradeLog, logManager, notifier, upgradeErr)
		return upgradeErr
	}
	logs.Println("[UPDATE]current version backed up to:", backupPath)
	upgradeLog.BackupPath = backupPath
	_ = logManager.UpdateLog(upgradeLog)
	// ??????7????
	cleaner.AddFileWithDelay(backupPath, 7*24*time.Hour)

	notifier.NotifyProgress("admin", 0.9, "Replacing binary...")

	// 閺囨寧宕叉禍宀冪箻閸掕埖鏋冩禒?
	err = os.Chmod(binaryPath, 0755)
	if err != nil {
		upgradeErr := utils.NewUpgradeError(utils.StageInstall, utils.ErrCodePermissionDenied,
			"chmod new binary failed", err)
		handleUpgradeError(upgradeLog, logManager, notifier, upgradeErr)
		return upgradeErr
	}

	// 閸忓牆鐨剧拠鏇犳纯閹恒儴顩惄?
	err = copyFile(binaryPath, currentBinary)
	if err != nil {
		upgradeErr := utils.NewUpgradeError(utils.StageInstall, utils.ErrCodeInstallFailed,
			"replace binary failed", err)
		handleUpgradeError(upgradeLog, logManager, notifier, upgradeErr)
		return upgradeErr
	}

	logs.Println("[UPDATE]binary updated successfully")

	// 閺囧瓨鏌妛eb閻╊喖缍嶉敍鍫濐洤閺嬫粌鐡ㄩ崷顭掔礆
	webSrcDir := filepath.Join(extractDir, "web")
	if _, err := os.Stat(webSrcDir); err == nil {
		webDestDir := Tea.Root + "/web"
		_ = os.RemoveAll(webDestDir)
		err = copyDir(webSrcDir, webDestDir)
		if err != nil {
			logs.Println("[UPDATE]web update failed:", err.Error())
		} else {
			logs.Println("[UPDATE]web directory updated")
		}
	}

	notifier.NotifyProgress("admin", 0.95, "Update complete, restarting...")

	// 閺囧瓨鏌婇幋鎰
	duration := time.Since(startTime)
	upgradeLog.Status = utils.StatusSuccess
	upgradeLog.EndTime = time.Now()
	upgradeLog.DownloadSpeed = float64(upgradeLog.DownloadSize) / duration.Seconds() / 1024 / 1024
	_ = logManager.UpdateLog(upgradeLog)

	logs.Println("[UPDATE]update completed successfully, version:", version)
	logs.Println("[UPDATE]restarting service...")

	notifier.NotifySuccess("admin", version, duration)

	// 闁插秴鎯庨張宥呭
	return restartService()
}

// handleUpgradeError 婢跺嫮鎮婇崡鍥╅獓闁挎瑨顕?
func handleUpgradeError(log *utils.UpgradeLog, logManager *utils.UpgradeLogManager,
	notifier utils.UpdateNotifier, err *utils.UpgradeError) {
	log.Status = utils.StatusFailed
	log.ErrorCode = int(err.Code)
	log.ErrorMessage = err.Message
	log.ErrorStage = string(err.Stage)
	log.EndTime = time.Now()
	_ = logManager.UpdateLog(log)

	notifier.NotifyFailed(log.Component, log.NewVersion, err)
}

// downloadFileWithProgress 娑撳娴囬弬鍥︽楠炶埖妯夌粈楦跨箻鎼?
func downloadFileWithProgress(client *http.Client, url, dest string, progressCallback func(progress float32, speed float64)) error {
	if client == nil {
		client = downloadHTTPClient
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("http status: %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	contentLength := resp.ContentLength
	hasContentLength := contentLength > 0
	downloaded := int64(0)
	startTime := time.Now()
	lastNotifyTime := startTime

	buffer := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := out.Write(buffer[:n])
			if writeErr != nil {
				return writeErr
			}

			downloaded += int64(n)

			// download progress update; speed still reported when size unknown
			if time.Since(lastNotifyTime) >= time.Second {
				speed := float64(downloaded) / time.Since(startTime).Seconds() / 1024 / 1024
				if hasContentLength {
					progress := float32(downloaded) / float32(contentLength)
					progressCallback(progress, speed)
				} else {
					progressCallback(-1, speed)
				}
				lastNotifyTime = time.Now()
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	// final speed/progress update
	speed := float64(downloaded) / time.Since(startTime).Seconds() / 1024 / 1024
	if hasContentLength {
		progressCallback(1.0, speed)
	} else {
		progressCallback(-1, speed)
	}

	return nil
}
func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// 閸氬本顒為崚鎵梿閻?
	err = destFile.Sync()
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func restartService() error {
	// 灏濊瘯浣跨敤systemctl閲嶅惎
	cmd := exec.Command("systemctl", "restart", teaconst.SystemdServiceName)
	err := cmd.Run()
	if err == nil {
		return nil
	}

	// 濡傛灉systemctl澶辫触锛屽垯鐩存帴鎷夎捣鏂拌繘绋嬪苟閫€鍑哄綋鍓嶈繘绋?
	logs.Println("[UPDATE]systemctl restart failed, trying direct restart")

	exePath, pathErr := os.Executable()
	if pathErr != nil {
		return fmt.Errorf("locate executable failed: %w", pathErr)
	}

	newCmd := exec.Command(exePath, os.Args[1:]...)
	newCmd.Stdout = os.Stdout
	newCmd.Stderr = os.Stderr
	newCmd.Env = os.Environ()
	if startErr := newCmd.Start(); startErr != nil {
		return fmt.Errorf("start new process failed: %w", startErr)
	}

	// 寤惰繜閫€鍑猴紝淇濊瘉褰撳墠璇锋眰瀹屾垚
	time.AfterFunc(1*time.Second, func() {
		os.Exit(0)
	})

	return nil
}
