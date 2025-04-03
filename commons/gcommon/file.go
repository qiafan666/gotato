package gcommon

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type FileReader struct {
	*bufio.Reader
	file   *os.File
	offset int64
}

// NewFileReader 创建文件读取器
func NewFileReader(path string) (*FileReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &FileReader{
		file:   file,
		Reader: bufio.NewReader(file),
		offset: 0,
	}, nil
}

// ReadLine 读取一行数据，直到遇到换行符或文件结束
func (f *FileReader) ReadLine() (string, error) {
	data, err := f.Reader.ReadBytes('\n')
	f.offset += int64(len(data))
	if err == nil || err == io.EOF {
		for len(data) > 0 && (data[len(data)-1] == '\r' || data[len(data)-1] == '\n') {
			data = data[:len(data)-1]
		}
		return string(data), err
	}
	return "", err
}

// Offset 返回读取偏移量
func (f *FileReader) Offset() int64 {
	return f.offset
}

// SetOffset 设置偏移量
func (f *FileReader) SetOffset(offset int64) error {
	_, err := f.file.Seek(offset, 0)
	if err != nil {
		return err
	}
	f.Reader = bufio.NewReader(f.file)
	f.offset = offset
	return nil
}

// Close 关闭文件读取
func (f *FileReader) Close() error {
	return f.file.Close()
}

// FileIsExist 检查文件夹或文件路径是否存在
func FileIsExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}

// FileCreate 创建路径文件
func FileCreate(path string) bool {
	file, err := os.Create(path)
	if err != nil {
		return false
	}

	defer file.Close()
	return true
}

// DirCreate 创建路径，不存在的情况下创建目录，已存在的层级跳过
func DirCreate(absPath string) error {
	// return os.MkdirAll(path.Dir(absPath), os.ModePerm)
	return os.MkdirAll(absPath, os.ModePerm)
}

// DirCopy 复制目录
func DirCopy(srcPath string, dstPath string) error {
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to get source directory info: %w", err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("source path is not a directory: %s", srcPath)
	}

	err = os.MkdirAll(dstPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	entries, err := os.ReadDir(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		srcDir := filepath.Join(srcPath, entry.Name())
		dstDir := filepath.Join(dstPath, entry.Name())

		if entry.IsDir() {
			err = DirCopy(srcDir, dstDir)
			if err != nil {
				return err
			}
		} else {
			err = FileCopy(srcDir, dstDir)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DirIsExist 检查是否为目录
func DirIsExist(path string) bool {
	file, err := os.Stat(path)
	if err != nil {
		return false
	}
	return file.IsDir()
}

// FileRemove 删除文件
func FileRemove(path string) error {
	return os.Remove(path)
}

// FileCopy 复制文件
func FileCopy(srcPath string, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	distFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer distFile.Close()

	var tmp = make([]byte, 1024*4)
	for {
		n, err := srcFile.Read(tmp)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		_, err = distFile.Write(tmp[:n])
		if err != nil {
			return err
		}
	}
}

// FileClear 清空文件内容
func FileClear(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString("")
	return err
}

// FileReadToString 读取文件内容到字符串
func FileReadToString(path string) (string, error) {
	readBytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(readBytes), nil
}

// FileReadByLine 读取文件内容按行读取
func FileReadByLine(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		result = append(result, scanner.Text()) // 读取完整的一行
	}

	if err = scanner.Err(); err != nil {
		return nil, err // 返回读取过程中发生的错误
	}

	return result, nil
}

// FileListNames 返回目录下的文件名列表
func FileListNames(path string) ([]string, error) {
	if !FileIsExist(path) {
		return []string{}, nil
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return []string{}, err
	}

	sz := len(files)
	if sz == 0 {
		return []string{}, nil
	}

	result := []string{}
	for i := 0; i < sz; i++ {
		if !files[i].IsDir() {
			result = append(result, files[i].Name())
		}
	}

	return result, nil
}

// FileIsZip 检查是否为zip文件
func FileIsZip(filepath string) bool {
	file, err := os.Open(filepath)
	if err != nil {
		return false
	}
	defer file.Close()

	buf := make([]byte, 4)
	if n, err := file.Read(buf); err != nil || n < 4 {
		return false
	}

	return bytes.Equal(buf, []byte("PK\x03\x04"))
}

// FileZip 创建zip文件
func FileZip(path string, destPath string) error {
	if DirIsExist(path) {
		return zipFolder(path, destPath)
	}

	return zipFile(path, destPath)
}

func zipFile(filePath string, destPath string) error {
	zipFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	return addFileToArchive1(filePath, archive)
}

func zipFolder(folderPath string, destPath string) error {
	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	w := zip.NewWriter(outFile)

	err = addFileToArchive2(w, folderPath, "")
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return nil
}

func addFileToArchive1(fpath string, archive *zip.Writer) error {
	err := filepath.Walk(fpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, filepath.Dir(fpath)+"/")

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(writer, file); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func addFileToArchive2(w *zip.Writer, basePath, baseInZip string) error {
	files, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(basePath, "/") {
		basePath = basePath + "/"
	}

	for _, file := range files {
		if !file.IsDir() {
			dat, err := os.ReadFile(basePath + file.Name())
			if err != nil {
				return err
			}

			f, err := w.Create(baseInZip + file.Name())
			if err != nil {
				return err
			}
			_, err = f.Write(dat)
			if err != nil {
				return err
			}
		} else if file.IsDir() {
			newBase := basePath + file.Name() + "/"
			addFileToArchive2(w, newBase, baseInZip+file.Name()+"/")
		}
	}

	return nil
}

// FileUnZip 解压zip文件
func FileUnZip(zipFile string, destPath string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		decodeName := f.Name
		if f.Flags == 0 {
			i := bytes.NewReader([]byte(f.Name))
			decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
			content, _ := io.ReadAll(decoder)
			decodeName = string(content)
		}
		// issue#62: fix ZipSlip bug
		var path string
		path, err = safeFilepathJoin(destPath, decodeName)
		if err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
			if err != nil {
				return err
			}

			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()

			outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ZipAppendEntry 拼接zip文件
func ZipAppendEntry(fpath string, destPath string) error {
	tempFile, err := os.CreateTemp("", "temp.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	zipReader, err := zip.OpenReader(destPath)
	if err != nil {
		return err
	}

	archive := zip.NewWriter(tempFile)

	for _, zipItem := range zipReader.File {
		zipItemReader, err := zipItem.Open()
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(zipItem.FileInfo())
		if err != nil {
			return err
		}
		header.Name = zipItem.Name
		targetItem, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(targetItem, zipItemReader)
		if err != nil {
			return err
		}
	}

	err = addFileToArchive1(fpath, archive)

	if err != nil {
		return err
	}

	err = zipReader.Close()
	if err != nil {
		return err
	}
	err = archive.Close()
	if err != nil {
		return err
	}
	err = tempFile.Close()
	if err != nil {
		return err
	}

	return FileCopy(tempFile.Name(), destPath)
}

func safeFilepathJoin(path1, path2 string) (string, error) {
	relPath, err := filepath.Rel(".", path2)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("(zipslip) filepath is unsafe %q: %v", path2, err)
	}
	if path1 == "" {
		path1 = "."
	}
	return filepath.Join(path1, filepath.Join("/", relPath)), nil
}

// FileIsLink 检查是否为软链接
func FileIsLink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// FileMode 返回文件权限
func FileMode(path string) (fs.FileMode, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}
	return fi.Mode(), nil
}

// FileMiMeType 返回文件媒体类型
func FileMiMeType(file any) string {
	var mediatype string

	readBuffer := func(f *os.File) ([]byte, error) {
		buffer := make([]byte, 512)
		_, err := f.Read(buffer)
		if err != nil {
			return nil, err
		}
		return buffer, nil
	}

	if filePath, ok := file.(string); ok {
		f, err := os.Open(filePath)
		if err != nil {
			return mediatype
		}
		buffer, err := readBuffer(f)
		if err != nil {
			return mediatype
		}
		return http.DetectContentType(buffer)
	}

	if f, ok := file.(*os.File); ok {
		buffer, err := readBuffer(f)
		if err != nil {
			return mediatype
		}
		return http.DetectContentType(buffer)
	}
	return mediatype
}

// CurrentPath 返回当前执行文件的路径
func CurrentPath() string {
	var absPath string
	_, filename, _, ok := runtime.Caller(1)
	if ok {
		absPath = filepath.Dir(filename)
	}

	return absPath
}

// FileSize 返回文件大小
func FileSize(path string) (int64, error) {
	f, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return f.Size(), nil
}

// DirSize 返回目录大小
func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			size += info.Size()
		}
		return err
	})
	return size, err
}

// FileUpdateTime 返回文件修改时间
func FileUpdateTime(filepath string) (int64, error) {
	f, err := os.Stat(filepath)
	if err != nil {
		return 0, err
	}
	return f.ModTime().Unix(), nil
}

// FileSha 返回文件sha值，shaType为1,256,512
func FileSha(filepath string, shaType ...int) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha1.New()
	if len(shaType) > 0 {
		if shaType[0] == 1 {
			h = sha1.New()
		} else if shaType[0] == 256 {
			h = sha256.New()
		} else if shaType[0] == 512 {
			h = sha512.New()
		} else {
			return "", errors.New("param `shaType` should be 1, 256 or 512")
		}
	}

	_, err = io.Copy(h, file)

	if err != nil {
		return "", err
	}

	sha := fmt.Sprintf("%x", h.Sum(nil))

	return sha, nil

}

// FileReadCsv 读取csv文件内容
func FileReadCsv(filepath string, delimiter ...rune) ([][]string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	if len(delimiter) > 0 {
		reader.Comma = delimiter[0]
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

// FileWriteCsv 写入csv文件内容
// append: 拼接到已有文件中
func FileWriteCsv(filepath string, records [][]string, append bool, delimiter ...rune) error {
	flag := os.O_RDWR | os.O_CREATE

	if append {
		flag = flag | os.O_APPEND
	}

	f, err := os.OpenFile(filepath, flag, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	writer := csv.NewWriter(f)
	// 设置默认分隔符为逗号，除非另外指定
	if len(delimiter) > 0 {
		writer.Comma = delimiter[0]
	} else {
		writer.Comma = ','
	}

	// 遍历所有记录并处理包含分隔符或双引号的单元格
	for i := range records {
		for j := range records[i] {
			records[i][j] = escapeCSVField(records[i][j], writer.Comma)
		}
	}

	return writer.WriteAll(records)
}

// FileWriteString 写入字符串到文件
func FileWriteString(filepath string, content string, append bool) error {
	var flag int
	if append {
		flag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	} else {
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	}

	f, err := os.OpenFile(filepath, flag, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}

// FileWriteBytes 写入字节到文件
func FileWriteBytes(filepath string, content []byte) error {
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(content)
	return err
}

// FileRead 读取文件内容，返回io.ReadCloser, 关闭函数
func FileRead(path string) (reader io.ReadCloser, closeFn func(), err error) {
	if FileIsExist(path) {
		reader, err = os.Open(path)
		if err != nil {
			return nil, func() {}, err
		}
		return reader, func() { reader.Close() }, nil
	} else {
		return nil, func() {}, errors.New("unknown file type")
	}
}

// escapeCSVField 处理单元格内容，如果包含分隔符，则用双引号包裹
func escapeCSVField(field string, delimiter rune) string {
	// 替换所有的双引号为两个双引号
	escapedField := strings.ReplaceAll(field, "\"", "\"\"")

	// 如果字段包含分隔符、双引号或换行符，用双引号包裹整个字段
	if strings.ContainsAny(escapedField, string(delimiter)+"\"\n") {
		escapedField = fmt.Sprintf("\"%s\"", escapedField)
	}

	return escapedField
}

// FileWriteMapsToCsv 写入map到csv文件
// filepath: csv文件路径
// records: map切片，每个map表示一行数据
// appendToExistingFile: 是否追加到已有文件中
// delimiter: csv文件分隔符，默认逗号
// headers: csv文件头部，如果为空，则使用map的key作为头部
func FileWriteMapsToCsv(filepath string, records []map[string]any, appendToExistingFile bool, delimiter rune,
	headers ...[]string) error {
	for _, record := range records {
		for _, value := range record {
			if !isCsvSupportedType(value) {
				return errors.New("unsupported value type detected; only basic types are supported: \nbool, rune, string, int, int64, float32, float64, uint, byte, complex128, complex64, uintptr")
			}
		}
	}

	var columnHeaders []string
	if len(headers) > 0 {
		columnHeaders = headers[0]
	} else {
		columnHeaders = make([]string, 0, len(records[0]))
		for key := range records[0] {
			columnHeaders = append(columnHeaders, key)
		}
		// sort keys in alphabeta order
		sort.Strings(columnHeaders)
	}

	var datasToWrite [][]string
	if !appendToExistingFile {
		datasToWrite = append(datasToWrite, columnHeaders)
	}

	for _, record := range records {
		row := make([]string, 0, len(columnHeaders))
		for _, h := range columnHeaders {
			row = append(row, fmt.Sprintf("%v", record[h]))
		}
		datasToWrite = append(datasToWrite, row)
	}

	return FileWriteCsv(filepath, datasToWrite, appendToExistingFile, delimiter)
}

// isCsvSupportedType 检查是否为csv支持的类型
func isCsvSupportedType(v interface{}) bool {
	switch v.(type) {
	case bool, rune, string, int, int64, float32, float64, uint, byte, complex128, complex64, uintptr:
		return true
	default:
		return false
	}
}

// FileChunkRead 读取文件块，返回每行内容
func FileChunkRead(file *os.File, offset int64, size int, bufPool *sync.Pool) ([]string, error) {
	buf := bufPool.Get().([]byte)[:size] // 从Pool获取缓冲区并调整大小
	n, err := file.ReadAt(buf, offset)   // 从指定偏移读取数据到缓冲区
	if err != nil && err != io.EOF {
		return nil, err
	}
	buf = buf[:n] // 调整切片以匹配实际读取的字节数

	var lines []string
	var lineStart int
	for i, b := range buf {
		if b == '\n' {
			line := string(buf[lineStart:i]) // 不包括换行符
			lines = append(lines, line)
			lineStart = i + 1 // 设置下一行的开始
		}
	}

	if lineStart < len(buf) { // 处理块末尾的行
		line := string(buf[lineStart:])
		lines = append(lines, line)
	}
	bufPool.Put(buf) // 读取完成后，将缓冲区放回Pool
	return lines, nil
}

// FileParallelChunkRead 读取文件块并并发处理，返回每行内容
// filePath 文件路径
// chunkSizeMB 分块的大小（单位MB，设置为0时使用默认100MB）,设置过大反而不利，视情调整
// maxGoroutine 并发读取分块的数量，设置为0时使用CPU核心数
// linesCh用于接收返回结果的通道。
func FileParallelChunkRead(filePath string, linesCh chan<- []string, chunkSizeMB, maxGoroutine int) error {
	if chunkSizeMB == 0 {
		chunkSizeMB = 100
	}
	chunkSize := chunkSizeMB * 1024 * 1024
	// 内存复用
	bufPool := sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, chunkSize)
		},
	}

	if maxGoroutine == 0 {
		maxGoroutine = runtime.NumCPU() // 设置为0时使用CPU核心数
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	chunkOffsetCh := make(chan int64, maxGoroutine)

	// 分配工作
	go func() {
		for i := int64(0); i < info.Size(); i += int64(chunkSize) {
			chunkOffsetCh <- i
		}
		close(chunkOffsetCh)
	}()

	// 启动工作协程
	for i := 0; i < maxGoroutine; i++ {
		wg.Add(1)
		go func() {
			for chunkOffset := range chunkOffsetCh {
				chunk, err := FileChunkRead(f, chunkOffset, chunkSize, &bufPool)
				if err == nil {
					linesCh <- chunk
				}

			}
			wg.Done()
		}()
	}

	// 等待所有解析完成后关闭行通道
	wg.Wait()
	close(linesCh)

	return nil
}
