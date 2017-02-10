package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"time"
)

func readUpdateList() (map[string]string, error) {
	result := make(map[string]string)

	r, err := regexp.Compile(`__(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}).log$`)
	if err != nil {
		return result, err
	}

	logPath := path.Join(updateDirPath, "log")
	files, err := ioutil.ReadDir(logPath)
	if err != nil {
		return result, err
	}

	for _, f := range files {
		if match := r.FindStringSubmatch(f.Name()); len(match) == 2 {
			result[match[1]] = f.Name()
		}
	}

	return result, nil
}

func getUpdateListHandler(w http.ResponseWriter, req *http.Request) {
	updateList, err := readUpdateList()
	if err != nil {
		returnError(w, req, http.StatusInternalServerError, "Cannot read list of updates", err)
		return
	}

	keys := make([]string, 0, len(updateList))
	for k := range updateList {
		keys = append(keys, k)
	}

	json.NewEncoder(w).Encode(keys)
}

func getUpdateHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	returnLog(w, req, params["timestamp"], http.StatusOK)
}

func returnLog(w http.ResponseWriter, req *http.Request, timestamp string, httpStatus int) {
	updateList, err := readUpdateList()
	if err != nil {
		returnError(w, req, http.StatusInternalServerError, "Cannot read list of updates", err)
		return
	}

	logFile := updateList[timestamp]
	if logFile == "" {
		returnError(w, req, http.StatusNotFound, "Cannot find update", nil)
		return
	}

	filename := path.Join(updateDirPath, "log", logFile)
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		returnError(w, req, http.StatusInternalServerError, "Cannot read update log", err)
		return
	}

	re, err := regexp.Compile(`PID: (\d+)`)
	if err != nil {
		returnError(w, req, http.StatusInternalServerError, "Cannot find PID in update log", err)
		return
	}

	match := re.FindStringSubmatch(string(fileContent))
	if len(match) != 2 {
		returnError(w, req, http.StatusInternalServerError, "Cannot find PID in update log", nil)
		return
	}

	pidInt, err := strconv.Atoi(match[1])
	if err != nil {
		returnError(w, req, http.StatusInternalServerError, "Cannot find PID in update log", nil)
		return
	}

	updateState := "finished"
	if isPidAlive(pidInt) {
		updateState = "in-progress"
	}

	location := fmt.Sprintf("http://%s%s/v1/updates/%s", req.Host, pathPrefix, timestamp)
	w.Header().Set("Location", location)
	w.WriteHeader(httpStatus)

	json.NewEncoder(w).Encode(jsonResponce{
		Code:   httpStatus,
		Status: http.StatusText(httpStatus),
		Title:  updateState,
		Detail: string(fileContent),
	})
}

func runUpdateHandler(w http.ResponseWriter, req *http.Request) {
	if err := exec.Command("screen", "-d", "-m", "/usr/bin/pmm-update").Run(); err != nil {
		returnError(w, req, http.StatusInternalServerError, "Cannot run update", err)
		return
	}

	// Advanced Sleep Programming :)
	time.Sleep(1 * time.Second)

	timestamp, _, err := getCurrentUpdate()
	if timestamp == "" || err != nil {
		returnError(w, req, http.StatusInternalServerError, "Cannot find update log", err)
	}

	returnLog(w, req, timestamp, http.StatusAccepted)
}

func getCurrentUpdate() (string, int, error) {
	pidFile := path.Join(updateDirPath, "pmm-update.pid")
	pid, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return "", -1, err
	}

	pidStr := string(pid[:len(pid)-1])
	pidInt, err := strconv.Atoi(pidStr)
	if err != nil {
		return "", -1, err
	}

	pattern := fmt.Sprintf("PID: %s$", pidStr)
	logPath := path.Join(updateDirPath, "log/*.log")
	logs, err := filepath.Glob(logPath)
	if err != nil {
		return "", -1, err
	}

	args := append([]string{pattern}, logs...)
	currentLogOutput, err := exec.Command("grep", args...).Output()
	if err != nil {
		return "", -1, err
	}

	re, err := regexp.Compile(`__(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}).log:`)
	if err != nil {
		return "", -1, err
	}

	match := re.FindStringSubmatch(string(currentLogOutput))
	if len(match) != 2 {
		return "", -1, err
	}
	return match[1], pidInt, nil
}

func deleteUpdateHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	updateList, err := readUpdateList()
	if err != nil {
		returnError(w, req, http.StatusInternalServerError, "Cannot read list of updates", err)
		return
	}

	logFile := updateList[params["timestamp"]]
	if logFile == "" {
		returnError(w, req, http.StatusNotFound, "Cannot find update", nil)
		return
	}

	filename := path.Join(updateDirPath, "log", logFile)
	if err = os.Remove(filename); err != nil {
		returnError(w, req, http.StatusInternalServerError, "Cannot remove update log", nil)
		return
	}
	returnSuccess(w)
}