// +build linux

package pcap

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("pcap")

type File struct {
	location string
	fileTime time.Time
	size     int64
}

type Cleaner struct {
	maxDirectorySize    int64
	diskFreeSpaceMargin int64
	cleanPeriod         time.Duration
	pcapDataRetention   time.Duration
	baseDirectory       string

	fileLock *FileLock
}

func NewCleaner(cleanPeriod time.Duration, maxDirectorySize, diskFreeSpaceMargin int64, baseDirectory string) *Cleaner {
	return &Cleaner{
		maxDirectorySize:    maxDirectorySize,
		diskFreeSpaceMargin: diskFreeSpaceMargin,
		cleanPeriod:         cleanPeriod,
		baseDirectory:       baseDirectory,
		fileLock:            New(baseDirectory),
	}
}

func (c *Cleaner) UpdatePcapDataRetention(pcapDataRetention time.Duration) {
	atomic.StoreInt64((*int64)(&c.pcapDataRetention), int64(pcapDataRetention))
}

func (c *Cleaner) GetPcapDataRetention() time.Duration {
	return time.Duration(atomic.LoadInt64((*int64)(&c.pcapDataRetention)))
}

func (c *Cleaner) work() {
	var files []File
	for now := range time.Tick(c.cleanPeriod) {
		c.fileLock.Lock()
		files = files[:0]
		filepath.Walk(c.baseDirectory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Debugf("Walk directory error: %s", err)
				// 返回nil，否则Walk()会中止
				return nil
			}
			name := info.Name()
			if info.IsDir() || !strings.HasSuffix(name, ".pcap") {
				return nil
			}
			files = append(files, File{
				location: path,
				fileTime: info.ModTime(),
				size:     info.Size(),
			})
			return nil
		})
		// 用结束写入时间倒排
		sort.Slice(files, func(i, j int) bool { return files[i].fileTime.Sub(files[j].fileTime) > 0 })

		// check delete
		sumSize := int64(0)
		nDeleted := 0
		pcapDataRetention := c.GetPcapDataRetention()
		for _, f := range files {
			sumSize += f.size
			if sumSize >= c.maxDirectorySize || (pcapDataRetention != 0 && now.Sub(f.fileTime) > pcapDataRetention) {
				os.Remove(f.location)
				nDeleted++
			}
		}

		fs := syscall.Statfs_t{}
		err := syscall.Statfs(c.baseDirectory, &fs)
		if err == nil {
			free := int64(fs.Bfree) * int64(fs.Bsize)
			for i := len(files) - nDeleted - 1; i >= 0 && free < c.diskFreeSpaceMargin; i-- {
				os.Remove(files[i].location)
				free += files[i].size
			}
		}
		c.fileLock.Unlock()
	}
}

func (c *Cleaner) Start() {
	go c.work()
}
