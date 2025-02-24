package utils

import (
	"fmt"
	"github.com/fatih/color"
)

var logoColor = color.New(color.FgGreen).SprintFunc()
var bannerColor = color.New(color.FgHiYellow).SprintFunc()

func Banner() {
	logo := `
    ______                 ____                  __    _
   / ____/___ ________  __/ __ )____ _________  / /   (_)___  ___
  / __/ / __ '/ ___/ / / / __  / __ '/ ___/ _ \/ /   / / __ \/ _ \
 / /___/ /_/ (__  ) /_/ / /_/ / /_/ (__  )  __/ /___/ / / / /  __/
/_____/\__,_/____/\__, /_____/\__,_/____/\___/_____/_/_/ /_/\___/
                 /____/`
	banner := `
  version: 0.1
  author: zhtty
  github: https://github.com/zhan741965531/EasyBaseLine
`
	fmt.Println(logoColor(logo))
	fmt.Println(bannerColor(banner))
}
