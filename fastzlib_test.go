package czlib

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

var gzipped, raw []byte

func zip(b []byte) []byte {
	var out bytes.Buffer
	w := zlib.NewWriter(&out)
	w.Write(b)
	w.Close()
	return out.Bytes()
}

func init() {
	var err error
	payload := os.Getenv("PAYLOAD")
	if len(payload) == 0 {
		fmt.Println("You must provide PAYLOAD env var for path to test payload.")
		return
	}
	raw, err = ioutil.ReadFile(payload)
	if err != nil {
		fmt.Printf("Error opening payload: %s\n", err)
	}
	gzipped = zip(raw)
	// fmt.Printf("%d byte test payload (%d orig)\n", len(gzipped), len(raw))
}

// Generate an n-byte long []byte
func genData(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

func TestAllZlib(t *testing.T) {
	type compressFunc func([]byte) ([]byte, error)
	funcs := []compressFunc{Compress, gzip, zzip}
	names := []string{"Compress", "gzip", "zzip"}
	for _, i := range []int{10, 128, 1000, 1024 * 10, 1024 * 100, 1024 * 1024, 1024 * 1024 * 7} {
		data, err := genData(i)
		if err != nil {
			t.Error(err)
			continue
		}
		for i, f := range funcs {
			comp, err := f(data)
			if err != nil {
				t.Fatalf("Compression failed on %v: %s", names[i], err)
			}
			decomp, err := Decompress(comp)
			if err != nil {
				t.Fatalf("Decompression failed on %v: %s", names[i], err)
			}
			if bytes.Compare(decomp, data) != 0 {
				t.Fatalf("deflate->inflate does not match original for %s", names[i])
			}
		}
	}
}

func TestEmpty(t *testing.T) {
	var empty []byte
	_, err := Compress(empty)
	if err != nil {
		t.Fatalf("unexpected error compressing empty slice")
	}
	_, err = Decompress(empty)
	if err == nil {
		t.Fatalf("unexpected success decompressing empty slice")
	}
}

func TestUnsafeZlib(t *testing.T) {
	for _, i := range []int{10, 128, 1000, 1024 * 10, 1024 * 100, 1024 * 1024, 1024 * 1024 * 7} {
		data, err := genData(i)
		if err != nil {
			t.Error(err)
			continue
		}
		comp, err := UnsafeCompress(data)
		if err != nil {
			t.Fatal(err)
		}
		decomp, err := UnsafeDecompress(comp)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Compare(decomp, data) != 0 {
			t.Fatal("Compress -> Decompress on byte array failed to match original data.")
		}
		comp.Free()
		decomp.Free()
	}
}

// Compression benchmarks
func BenchmarkCompressUnsafe(b *testing.B) {
	if raw == nil {
		b.Skip("You must provide PAYLOAD env var for benchmarking.")
	}
	b.SetBytes(int64(len(raw)))
	for i := 0; i < b.N; i++ {
		u, _ := UnsafeCompress(raw)
		u.Free()
	}
}

func BenchmarkCompress(b *testing.B) {
	if raw == nil {
		b.Skip("You must provide PAYLOAD env var for benchmarking.")
	}
	b.SetBytes(int64(len(raw)))
	for i := 0; i < b.N; i++ {
		Compress(raw)
	}
}

func BenchmarkCompressStream(b *testing.B) {
	if raw == nil {
		b.Skip("You must provide PAYLOAD env var for benchmarking.")
	}
	b.SetBytes(int64(len(raw)))
	for i := 0; i < b.N; i++ {
		gzip(raw)
	}
}

func BenchmarkCompressStdZlib(b *testing.B) {
	if raw == nil {
		b.Skip("You must provide PAYLOAD env var for benchmarking.")
	}
	b.SetBytes(int64(len(raw)))
	for i := 0; i < b.N; i++ {
		zzip(raw)
	}
}

// Decomression benchmarks

func BenchmarkDecompressUnsafe(b *testing.B) {
	if raw == nil {
		b.Skip("You must provide PAYLOAD env var for benchmarking.")
	}
	b.SetBytes(int64(len(raw)))
	for i := 0; i < b.N; i++ {
		u, _ := UnsafeDecompress(gzipped)
		u.Free()
	}
}

func BenchmarkDecompress(b *testing.B) {
	if raw == nil {
		b.Skip("You must provide PAYLOAD env var for benchmarking.")
	}
	b.SetBytes(int64(len(raw)))
	for i := 0; i < b.N; i++ {
		Decompress(gzipped)
	}
}

func BenchmarkDecompressStream(b *testing.B) {
	if raw == nil {
		b.Skip("You must provide PAYLOAD env var for benchmarking.")
	}
	b.SetBytes(int64(len(raw)))
	for i := 0; i < b.N; i++ {
		gunzip(gzipped)
	}
}

func BenchmarkDecompressStdZlib(b *testing.B) {
	if raw == nil {
		b.Skip("You must provide PAYLOAD env var for benchmarking.")
	}
	b.SetBytes(int64(len(raw)))
	for i := 0; i < b.N; i++ {
		zunzip(gzipped)
	}
}

// helpers

func gunzip(body []byte) ([]byte, error) {
	reader, err := NewReader(bytes.NewBuffer(body))
	if err != nil {
		return []byte{}, err
	}
	return ioutil.ReadAll(reader)
}

// unzip a gzipped []byte payload, returning the unzipped []byte and error
func zunzip(body []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewBuffer(body))
	if err != nil {
		return []byte{}, err
	}
	return ioutil.ReadAll(reader)
}

func gzip(body []byte) ([]byte, error) {
	outb := make([]byte, 0, 16*1024)
	out := bytes.NewBuffer(outb)
	writer := NewWriter(out)
	n, err := writer.Write(body)
	if n != len(body) {
		return []byte{}, fmt.Errorf("compressed %d, expected %d", n, len(body))
	}
	if err != nil {
		return []byte{}, err
	}
	err = writer.Close()
	if err != nil {
		return []byte{}, err
	}
	return out.Bytes(), nil
}

func zzip(body []byte) ([]byte, error) {
	outb := make([]byte, 0, len(body))
	out := bytes.NewBuffer(outb)
	writer := zlib.NewWriter(out)
	n, err := writer.Write(body)
	if n != len(body) {
		return []byte{}, fmt.Errorf("compressed %d, expected %d", n, len(body))
	}
	if err != nil {
		return []byte{}, err
	}
	err = writer.Close()
	if err != nil {
		return []byte{}, err
	}
	return out.Bytes(), nil
}
