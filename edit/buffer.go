// Package edit provides an API for ed-like editing of file-backed buffers.
package edit

import (
	"encoding/binary"

	"github.com/eaburns/T/edit/buffer"
)

const (
	// BlockBytes is the block size used for the underlying buffer.Buffer.
	blockBytes = 1024
	// BlockRunes is the number of runes that fit in a buffer.Buffer block.
	blockRunes = blockBytes / runeBytes
)

// A Buffer is an unbounded buffer of runes.
type Buffer struct {
	bytes *buffer.Buffer
}

// NewBuffer returns a new Buffer.
// The buffer caches no more than blockSize runes in memory.
func NewBuffer() *Buffer {
	return &Buffer{bytes: buffer.New(blockBytes)}
}

// Close closes the buffer, freeing its resources.
func (b *Buffer) Close() error {
	return b.bytes.Close()
}

// Size returns the number of runes in the Buffer.
func (b *Buffer) Size() int64 {
	return b.bytes.Size() / runeBytes
}

// All returns the Address that identifies the entirety of the Buffer.
func (b *Buffer) All() Address {
	return Address{0, b.Size()}
}

// End returns the Address of the empty string at the end of the Buffer.
func (b *Buffer) End() Address {
	sz := b.Size()
	return Address{sz, sz}
}

// Read returns the runes in the range of an Address in the buffer.
func (b *Buffer) Read(at Address) ([]rune, error) {
	if at.From < 0 || at.From > at.To || at.To > b.Size() {
		return nil, AddressError(at)
	}

	bs := make([]byte, at.byteSize())
	if _, err := b.bytes.ReadAt(bs, at.fromByte()); err != nil {
		return nil, err
	}

	rs := make([]rune, 0, at.Size())
	for len(bs) > 0 {
		r := rune(binary.LittleEndian.Uint32(bs))
		rs = append(rs, r)
		bs = bs[runeBytes:]
	}
	return rs, nil
}

// Write writes runes to the range of an Address in the buffer.
func (b *Buffer) Write(rs []rune, at Address) error {
	if at.From < 0 || at.From > at.To || at.To > b.Size() {
		return AddressError(at)
	}

	if _, err := b.bytes.Delete(at.byteSize(), at.fromByte()); err != nil {
		return err
	}

	bs := make([]byte, len(rs)*runeBytes)
	for i, r := range rs {
		binary.LittleEndian.PutUint32(bs[i*runeBytes:], uint32(r))
	}

	_, err := b.bytes.Insert(bs, at.fromByte())
	return err
}
