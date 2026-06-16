# go-volumes/interface

The shared **block-device contract** for the go-volumes family — Go package
`volume`. A fixed-size, random-access store that higher layers read and write
through, **decoupling where the bytes live** (a pooled file volume, an S3-backed
chunk store, an NBD export, a qcow2/raw image) **from the format layered on top**
(ext4, xfs, squashfs, an OCI image, …).

Zero dependencies — standard library only (`io`). `CGO_ENABLED=0`.

## The contract

```go
// Device is a fixed-size, random-access block device.
// Writes MAY be buffered; Sync durably commits them.
type Device interface {
	io.ReaderAt           // ReadAt(p []byte, off int64) (int, error)
	io.WriterAt           // WriteAt(p []byte, off int64) (int, error)
	Size() (int64, error)
	Sync() error
	io.Closer
}
```

`Device` is deliberately the **intersection of what a real backing already
offers**: [`go-volumes/pool`](https://github.com/go-volumes/pool)'s `*Volume`
satisfies it **verbatim**.

- **`ReadOnly`** — the read-only subset (a read-only file, a squashfs/iso9660
  blob, a read-only NBD export, an OCI image). Any `Device` also satisfies it.

### Optional capabilities

A consumer type-asserts for these and degrades gracefully when a device does not
implement one:

| Interface | Capability |
|---|---|
| `Truncater` | resize in place — `Truncate(size int64)` |
| `Named` | carries a stable identifier — `Name()` |
| `ReadOnlyReporter` | reports whether it rejects writes — `ReadOnly()` |
| `Discarder` | release (TRIM/unmap) a byte range so a thin/sparse backing reclaims space |

## Where it sits

```
 format     ext4 · xfs · squashfs · OCI image      (go-filesystems)
            ─────────── Device / ReadOnly ───────────   ← this contract
 backing    pool *Volume · s3 *Store · oci image · NBD · raw/qcow2
```

`go-volumes/pool`'s `*Volume` **is** a `Device`; the `s3` and `oci` repos are
pluggable **backings** that decide where a pool's bytes actually live. A
[`go-filesystems`](https://github.com/go-filesystems) driver writes its on-disk
image through a `Device`, never caring which backing is underneath.

BSD-3-Clause.
