// SensitiveCleaner - config
// 2021-02-04 19:13
// Benny <benny.think@gmail.com>

package main

import "os"

var Token = os.Getenv("TOKEN")

type userConfig map[string]int64
