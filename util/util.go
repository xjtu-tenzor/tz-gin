package util

import (
	"fmt"
	"os"
	"time"

	"github.com/logrusorgru/aurora/v3"
)

func ErrMsg(msg string) {
	aurora := aurora.NewAurora(true)
	fmt.Fprintf(os.Stderr, "%s", aurora.Red(msg))
}

func WarnMsg(msg string) {
	aurora := aurora.NewAurora(true)
	fmt.Printf("%s", aurora.Yellow(msg))
}

func SuccessMsg(msg string) {
	aurora := aurora.NewAurora(true)
	fmt.Printf("%s", aurora.Green(msg))
}

func Loading(stop chan int) {
	animationFrames := []string{"|", "/", "-", "\\"}

	// animationDuration := 3 * time.Second
	startTime := time.Now()
	output := os.Stdout

	select {
	case <-stop:
		return

	default:
		for {
			// 计算经过的时间
			elapsedTime := time.Since(startTime)

			// 计算当前帧索引
			frameIndex := int((elapsedTime.Seconds() / 0.1)) % len(animationFrames)

			// 清除上一帧并显示新的帧
			SuccessMsg(fmt.Sprintf("\rLoading %s", animationFrames[frameIndex]))

			// 刷新输出，使动画能够更新
			output.Sync()

			// 暂停一段时间以控制动画速度
			time.Sleep(100 * time.Millisecond)

			// 如果经过了指定的持续时间，退出循环
		}
	}
}
