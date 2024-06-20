package reader

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"dnscheck/logger"
	"dnscheck/tools"
)

type Reader struct {
	domainsFile   *os.File
	workspaceFile *os.File
	scanner       *bufio.Scanner
	workspace     string
	from          int
	to            int
	currentLine   int
	// batchSize is the number of lines that will be read at once in every call.
	batchSize int
}
type Data struct {
	Domains []string
	Hash    string
}

func (r *Reader) Close() {
	if r.domainsFile != nil {
		r.domainsFile.Close()
	}
	if r.workspaceFile != nil {
		r.workspaceFile.Close()
	}
}

func (r *Reader) Batch() ([]string, error) {

	// update last successfull state to the workspace
	err := r.updateWorkspace()
	if err != nil {
		return nil, err
	}

	if r.currentLine == r.to && r.to != -1 {
		return nil, ErrRangeReached
	}

	result := make([]string, 0)

	for i := 0; i < r.batchSize; i++ {
		if !r.scanner.Scan() {
			if len(result) > 0 {
				return result, nil
			}

			if e := r.scanner.Err(); e != nil {
				logger.Info("reading from file has been stopped, cause:", e)
			}

			return nil, ErrDomainsEOF
		}
		line := strings.TrimSpace(r.scanner.Text())
		if len(line) == 0 {
			i--
			continue
		}
		line = strings.ToLower(line)

		result = append(result, line)
		r.currentLine++

		if r.scanner.Err() != nil {
			return nil, r.scanner.Err()
		}
	}

	return result, nil
}

func New(cfg Config) (*Reader, error) {

	r := &Reader{
		batchSize: cfg.BatchSize,
	}

	err := r.initWorkspace()
	if err != nil {
		return nil, err
	}

	err = r.initDomains(path.Join(r.workspace, cfg.DomainsFileName))
	if err != nil {
		return nil, err
	}

	return r, err
}
func (r *Reader) updateWorkspace() error {
	r.workspaceFile.Truncate(0)
	r.workspaceFile.Seek(0, 0)

	_, err := r.workspaceFile.WriteString(strings.Join([]string{
		fmt.Sprint(r.from),
		fmt.Sprint(r.to),
		fmt.Sprint(r.currentLine)},
		"\n"))
	return err
}
func (r *Reader) initDomains(domainPath string) error {

	logger.Header("initializing domains")

	file, err := os.Open(domainPath)
	if err != nil {
		return err
	}
	r.domainsFile = file

	r.scanner = bufio.NewScanner(r.domainsFile)

	logger.Info("skipping to the start point")

	from := r.from

	for i := 0; i < from+r.currentLine; i++ {
		if !r.scanner.Scan() {
			break
		}
	}
	logger.Info("done")

	return nil
}

/*
initWorkspace reads the workspace folder and looks for the workspace.lock file;
if it finds the workspace.lock file, it will validate it and then load the data.
in case of any validation error or if it cannot find the workspace.lock file, it will create a new file.
*/
func (r *Reader) initWorkspace() (err error) {

	logger.Header("initializing workspace")

	fname := path.Join(r.workspace, ".workspace.lock")

	_, fs, err := tools.CreateOrOpenFile(fname)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = fs.Close()
		} else {
			r.workspaceFile = fs
		}

	}()

	fs.Seek(0, 0)
	scanner := bufio.NewScanner(fs)

	fromLine := ""
	toLine := ""
	currentLine := ""

	if scanner.Scan() {
		fromLine = scanner.Text()
	}
	if scanner.Scan() {
		toLine = scanner.Text()
	}
	if scanner.Scan() {
		currentLine = scanner.Text()
	}
	currentWorkspace := true

	if len(fromLine) == 0 || len(toLine) == 0 || len(currentLine) == 0 {
		currentWorkspace = false
	} else {

		from, err := strconv.Atoi(fromLine)
		if err != nil {
			fromLine = ""
		}
		to, err := strconv.Atoi(toLine)
		if err != nil {
			toLine = ""
		}
		current, err := strconv.Atoi(currentLine)
		if err != nil {
			currentLine = ""
		}
		if len(fromLine) == 0 || len(toLine) == 0 || len(currentLine) == 0 {
			logger.Error("workspace file has been changed manually.")
			fs.Truncate(0)
			fs.Seek(0, 0)
			currentWorkspace = false
		} else {
			if (from > to && to != -1) || current < from || (current > to && to != -1) {
				logger.Error("workspace file has been changed manually.")
				fs.Truncate(0)
				fs.Seek(0, 0)
				currentWorkspace = false
			} else {
				r.from = from
				r.to = to
				r.currentLine = current
				currentWorkspace = true
			}
		}
	}

	if currentWorkspace {
		logger.Warning("the previous workspace will be continued")
	} else {
		fs.WriteString(strings.Join([]string{
			fmt.Sprint(r.from),
			fmt.Sprint(r.to),
			fmt.Sprint(r.currentLine)},
			"\n"))
		logger.Warning("a new workspace has been created")
	}

	sT := "all"
	if r.to != -1 {
		sT = fmt.Sprint(r.to)
	}
	logger.Info("form: ", fmt.Sprint(r.from))
	logger.Info("to: ", sT)
	logger.Info("start: ", r.currentLine)

	return nil
}
