package file

import (
	"fmt"
	"io"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

type MultipartReadSeeker struct {
	db               db.Session
	blobTable        string
	fileID           int64
	totalSize        int64
	offset           int64
	parts            []FileBlobLite
	cumOffsets       []int64
	currentPartIndex int
	currentPartData  []byte
}

func (m *MultipartReadSeeker) Read(p []byte) (n int, err error) {
	if m.offset >= m.totalSize {
		return 0, io.EOF
	}

	// Find the part that contains the current offset
	partIndex := -1
	for i := range m.parts {
		if m.offset >= m.cumOffsets[i] && m.offset < m.cumOffsets[i]+m.parts[i].Size {
			partIndex = i
			break
		}
	}

	if partIndex == -1 {
		return 0, io.EOF
	}

	// If it's a different part, fetch it
	if m.currentPartIndex != partIndex {
		err := m.fetchPart(partIndex)
		if err != nil {
			return 0, err
		}
	}

	// Read from the current part
	relOffset := m.offset - m.cumOffsets[partIndex]
	n = copy(p, m.currentPartData[relOffset:])
	m.offset += int64(n)

	// If we need more data and we're not at the end
	if n < len(p) && m.offset < m.totalSize {
		n2, err2 := m.Read(p[n:])
		return n + n2, err2
	}

	return n, nil
}

func (m *MultipartReadSeeker) Seek(offset int64, whence int) (int64, error) {
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = m.offset + offset
	case io.SeekEnd:
		newOffset = m.totalSize + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}

	if newOffset < 0 {
		return 0, fmt.Errorf("negative seek offset")
	}
	// If newOffset is within the total size, accept it. If beyond, it depends on Read to return EOF.
	// Standard behavior allows seeking beyond end, but let's cap it or handle it in Read.
	m.offset = newOffset
	return m.offset, nil
}

func (m *MultipartReadSeeker) Close() error {
	m.currentPartData = nil
	return nil
}

func (m *MultipartReadSeeker) fetchPart(index int) error {
	part := m.parts[index]
	row, err := m.db.SQL().Select("blob").From(m.blobTable).Where(db.Cond{"id": part.ID}).QueryRow()
	if err != nil {
		return err
	}

	data := make([]byte, part.Size)
	err = row.Scan(&data)
	if err != nil {
		return err
	}

	m.currentPartData = data
	m.currentPartIndex = index
	return nil
}

func (f *FileOperations) getMultipartReadSeeker(file *dbmodels.FileMeta) (io.ReadSeeker, io.Closer, error) {
	parts := make([]FileBlobLite, 0)
	err := f.fileBlobTable().Find(db.Cond{"file_id": file.ID}).
		Select("id", "size", "part_id").
		OrderBy("part_id").
		All(&parts)
	if err != nil {
		return nil, nil, err
	}

	if len(parts) == 0 && file.Size > 0 {
		return nil, nil, fmt.Errorf("multipart file has no parts")
	}

	cumOffsets := make([]int64, len(parts))
	var currentOffset int64 = 0
	for i, part := range parts {
		cumOffsets[i] = currentOffset
		currentOffset += part.Size
	}

	seeker := &MultipartReadSeeker{
		db:               f.db,
		blobTable:        f.getBlobTableName(),
		fileID:           file.ID,
		totalSize:        file.Size,
		parts:            parts,
		cumOffsets:       cumOffsets,
		currentPartIndex: -1,
	}

	return seeker, seeker, nil
}
