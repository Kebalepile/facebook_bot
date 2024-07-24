package types

import (
	"sync"
)

// @description init instance of any bot that
// has method Init
type FacebookBot interface {
	Run(wg *sync.WaitGroup)
}
