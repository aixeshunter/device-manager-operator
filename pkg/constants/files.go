package constants

import (
	"bufio"
	"io"
	"os"
	"regexp"
)

func CreateDirectoryIfNotExist(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(path, 0777)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func WriteToFile(filePath string, output []byte) error {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	_, err = writer.Write(output)
	if err != nil {
		return err
	}
	_ = writer.Flush()
	return nil
}

func ReadFile(filePath string, match, target string, deleteEmpty bool) ([]byte, bool, error) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, false, err
	}

	defer f.Close()
	reader := bufio.NewReader(f)
	needHandle := false
	output := make([]byte, 0)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return output, needHandle, nil
			}
			return nil, needHandle, err
		}
		if deleteEmpty == true && len(line) == 0 {
			continue
		}

		if ok, _ := regexp.Match(match, line); ok {
			// reg := regexp.MustCompile(ORIGIN)
			// newByte := reg.ReplaceAllString(string(line), TARGET)
			output = append(output, []byte(target)...)
			if target != "" {
				output = append(output, []byte("\n")...)
			}
			if !needHandle {
				needHandle = true
			}
		} else {
			output = append(output, line...)
			output = append(output, []byte("\n")...)
		}
	}
}
