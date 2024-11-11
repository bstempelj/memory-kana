package utils

import "math/rand/v2"

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandHash(length int) string {
    b := make([]byte, length)
    for i := range b {
        b[i] = charset[rand.IntN(len(charset))]
    }
    return string(b)
}
