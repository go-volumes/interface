// Package volume defines the shared block-device contract for the go-volumes
// family: a fixed-size, random-access store that higher layers read and write
// through, decoupling WHERE the bytes live (a pooled file volume, an S3-backed
// chunk store, an NBD export, a qcow2/raw image) from the FORMAT layered on top
// (ext4, xfs, squashfs, an OCI image, …).
//
// The contract is deliberately the intersection of what a real backing already
// offers: github.com/go-volumes/pool's *Volume satisfies [Device] verbatim.
package volume

import "io"

// Device is a fixed-size, random-access block device. Writes MAY be buffered
// (a write-back cache, an S3-backed chunk store); Sync durably commits them.
//
// It is the read-write contract a filesystem-format driver writes its on-disk
// image through, and the shape an S3/NBD/pool backing exposes.
type Device interface {
	io.ReaderAt // ReadAt(p []byte, off int64) (n int, err error)
	io.WriterAt // WriteAt(p []byte, off int64) (n int, err error)

	// Size reports the device's current size in bytes.
	Size() (int64, error)

	// Sync durably commits any buffered writes.
	Sync() error

	io.Closer // Close releases the device; further use is undefined.
}

// ReadOnly is the read-only subset: a file opened read-only, a squashfs/iso9660
// blob, a read-only NBD export, an OCI image layer set. Any [Device] also
// satisfies it.
type ReadOnly interface {
	io.ReaderAt
	Size() (int64, error)
	io.Closer
}

// The interfaces below are OPTIONAL capabilities. A consumer type-asserts for
// them and degrades gracefully when a device does not implement one.

// Truncater is a [Device] whose size can be changed in place.
type Truncater interface {
	Truncate(size int64) error
}

// Named is a device that carries a stable identifier (e.g. a volume name).
type Named interface {
	Name() string
}

// ReadOnlyReporter is a device that reports whether it rejects writes.
type ReadOnlyReporter interface {
	ReadOnly() bool
}

// Discarder is a device that can release (TRIM/unmap) a byte range, letting a
// thin/sparse backing reclaim the space.
type Discarder interface {
	Discard(off, length int64) error
}
