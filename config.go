// SensitiveCleaner - config
// 2021-02-04 19:13
// Benny <benny.think@gmail.com>

package main

import "os"

var token = os.Getenv("TOKEN")
var redisHost = os.Getenv("REDIS")

type userConfig map[string]int64
