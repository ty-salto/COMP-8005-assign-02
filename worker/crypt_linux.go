// crypt_linux.go
//go:build linux

package main

/*
#cgo LDFLAGS: -lcrypt
#define _GNU_SOURCE
#include <crypt.h>
#include <stdlib.h>
#include <string.h>

static char* go_crypt_r(const char* key, const char* salt, struct crypt_data* data) {
    // crypt_data must be zeroed before first use
    return crypt_r(key, salt, data);
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

type cryptCtx struct {
	data C.struct_crypt_data
}

func newCryptCtx() *cryptCtx {
	var ctx cryptCtx
	// zero it (Go already zeros, but being explicit is fine)
	ctx.data.initialized = 0
	return &ctx
}

func cryptHashWithCtx(ctx *cryptCtx, candidate string, salt string) (string, error) {
	cKey := C.CString(candidate)
	cSalt := C.CString(salt)
	defer C.free(unsafe.Pointer(cKey))
	defer C.free(unsafe.Pointer(cSalt))

	out := C.go_crypt_r(cKey, cSalt, &ctx.data)
	if out == nil {
		return "", errors.New("crypt_r returned NULL")
	}
	return C.GoString(out), nil
}
