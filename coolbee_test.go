package coolbee

import (
	"fmt"
	"os"
	"regexp"
	"testing"
)

func TestCoolbee(t *testing.T) {
	regstr := regexp.MustCompile("<title>([^<]+)</title>")
	coolbee := Classic(os.Stdout)
	coolbee.Get("http://www.6renyou.com/", func(body RespBody) {
		if bodyStr, ok := body.(string); ok {
			titles := regstr.FindStringSubmatch(bodyStr)
			if len(titles) > 1 {
				fmt.Println("title : ", titles[1])
			}
		}
	})
	coolbee.Run()
}
