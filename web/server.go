package web

import (
	"EasyBaseLine/config"
	"EasyBaseLine/report"
	"EasyBaseLine/utils"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/ghodss/yaml"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	ip               string
	currentAccessKey string
	AuthUsername     string
	AuthPassword     string
	checkFileDir     string
	infoColor        = color.New(color.FgHiGreen).SprintFunc()
)

func init() {
	// Generate an access key initially
	RegenerateAccessKey()
}

func checkAuth(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false
	}

	// 解码 base64 凭据
	payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
	if err != nil {
		// 记录错误或根据需要处理它
		utils.Error(fmt.Sprintf("解码 base64 凭据出错: %v\n", err))
		return false
	}
	pair := strings.SplitN(string(payload), ":", 2)

	return len(pair) == 2 && pair[0] == AuthUsername && pair[1] == AuthPassword
}

func Server(ipString, checkfileDir string) {
	checkFileDir = checkfileDir
	ip = ipString
	AuthUsername, AuthPassword = utils.GenerateAccount()
	utils.Info(fmt.Sprintf("web凭证已生成: [%s:%s]\n", infoColor(AuthUsername), infoColor(AuthPassword)))

	checkItemList, err := ListObjectsInDirectory(checkFileDir)
	if err != nil {
		utils.Error(fmt.Sprintf("获取检查文件时出错:%v\n", err))
		return
	}

	deploymentInfoList := generateDeploymentInfoList(checkItemList)

	webServer(ipString, deploymentInfoList)
}

func webServer(ipString string, itemData []map[string]interface{}) {

	ip = ipString

	messages := make(chan string, 1)

	// Set a timer to regenerate the access key every hour
	go func() {
		for {
			<-time.After(30 * time.Minute)
			RegenerateAccessKey()
		}
	}()

	utils.Info(fmt.Sprintf("Server IP address: %s\n", infoColor(ip)))
	go httpFileServer(messages, ip, itemData)

	<-messages
}

func httpFileServer(done chan string, ip string, itemData []map[string]interface{}) {
	http.HandleFunc("/", downloadHandler)
	http.HandleFunc("/display", displayHandler(itemData))
	http.HandleFunc("/data", dataHandler)
	http.HandleFunc("/results", resultsHandler)
	// 启动文件服务器
	utils.Info(fmt.Sprintf("Listening and serving HTTP on :8080\n"))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		utils.Fatal(fmt.Sprintf("Failed to start file server: %v\n", err), color.New(color.FgHiRed).SprintFunc())
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	utils.Info(fmt.Sprintf("[%s] tried to access %s\n", infoColor(r.RemoteAddr), strings.ReplaceAll(r.URL.Path, "/", "")))
	keys, ok := r.URL.Query()["key"]
	if !ok || len(keys[0]) < 1 || keys[0] != currentAccessKey {
		//http.Error(w, "Forbidden", http.StatusForbidden)
		http.Redirect(w, r, "/display", http.StatusSeeOther)
		return
	}

	filePath := fmt.Sprintf("./config/scripts/%s", r.URL.Path[1:])

	// 检查文件是否存在
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 打开文件并读取内容
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 设置响应头，告知浏览器下载文件
	w.Header().Set("Content-Disposition", "attachment")

	// 将文件内容写入响应
	io.Copy(w, file)
}

func displayHandler(itemData []map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		utils.Info(fmt.Sprintf("[%s] viewed deployment commands\n", infoColor(strings.ReplaceAll(r.RemoteAddr, "/", ""))))
		w.Header().Set("Content-Type", "text/html")

		// 用于存储HTML部分的切片
		var htmlSections []string

		for id, item := range itemData {
			//替换密钥
			re, err := regexp.Compile(`key=[a-fA-F0-9]{16}`)
			if err != nil {
				utils.Warn(fmt.Sprintf("Error compiling regex:%s", err))
			}
			newKey := "key=" + currentAccessKey
			deploymentCommand := re.ReplaceAllString(item["deploymentCommand"].(string), newKey)
			// 生成HTML部分并添加到切片中
			idStr := strconv.Itoa(id)
			htmlSection := fmt.Sprintf(`
  <div class="section">
      <h2>%s:</h2>
      <button onclick="toggleInfo('%s')">检查信息</button>
      <div id="%sInfo" style="display: none;">
         <ul>
              <li><strong>检查 ID:</strong> %s</li>
              <li><strong>检查类型:</strong>  %s</li>
              <li><strong>检查描述:</strong>  %s</li>
              <li><strong>执行器:</strong>  %s</li>
              <li><strong>支持的操作系统:</strong>  %s</li>
              <li><strong>创建日期:</strong>  %s</li>
              <li><strong>最后修改日期:</strong>  %s</li>
              <li><strong>检查版本:</strong>  %f</li>
              <li><strong>额外信息:</strong>  %s</li>
          </ul>
      </div>
      <pre id="%s">%s</pre>
      <button onclick="copyToClipboard('%s')">复制</button>
  </div>`,
				item["check_name"],
				idStr,
				idStr,
				item["check_id"],
				item["check_type"],
				item["check_description"],
				item["check_executor"],
				strings.Join(item["operating_system"].([]string), ", "), // 强制转换为字符串切片
				item["creation_date"],
				item["last_modified_date"],
				item["check_version"],
				item["additional_information"],
				idStr,
				deploymentCommand,
				idStr)

			htmlSections = append(htmlSections, htmlSection)
		}

		index, _ := config.Content.ReadFile("html/index.html")

		// 将HTML部分切片合并到HTML模板中
		htmlOutput := fmt.Sprintf(string(index), strings.Join(htmlSections, "\n"))
		fmt.Fprintf(w, htmlOutput)
	}
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	// Check request method and content type
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		http.Error(w, "Invalid content type", http.StatusBadRequest)
		return
	}

	// Decode the JSON body
	var jsonResult config.CheckJsonResult
	err := json.NewDecoder(r.Body).Decode(&jsonResult)
	if err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("./config/json/%s_%s.json", jsonResult.HostIP, timestamp)
	file, err := os.Create(fileName)
	if err != nil {
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	} else {
		utils.Info(fmt.Sprintf("检查对象[%s]的[%s]检查结果已保存到:%s\n", infoColor(jsonResult.HostIP), infoColor(jsonResult.CheckId), fileName))
	}
	defer file.Close()

	report.GenerateReport(mathData(jsonResult))

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(jsonResult); err != nil {
		http.Error(w, "Error saving to file", http.StatusInternalServerError)
		return
	}

	// Respond back to the client
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status": "success",
		"hostIP": jsonResult.HostIP,
	}
	json.NewEncoder(w).Encode(response)
}

func resultsHandler(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(r) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}
	// 找到当前目录下的所有JSON文件
	files, err := ioutil.ReadDir("./config/json")
	if err != nil {
		http.Error(w, "Error reading directory", http.StatusInternalServerError)
		return
	}

	// 构建一个切片来保存检查结果
	var finalResults []config.FinalResult

	// 遍历所有JSON文件
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			filePath := filepath.Join("./config/json", file.Name())

			// 读取JSON文件
			jsonContent, err := ioutil.ReadFile(filePath)
			if err != nil {
				http.Error(w, "Error reading JSON file", http.StatusInternalServerError)
				return
			}

			// 解析JSON数据
			var jsonCheckResult config.CheckJsonResult
			if err := json.Unmarshal(jsonContent, &jsonCheckResult); err != nil {
				http.Error(w, "Error parsing JSON data", http.StatusInternalServerError)
				fmt.Println(err)
				return
			}

			finalResult := mathData(jsonCheckResult)
			finalResult.HostIP = jsonCheckResult.HostIP

			finalResults = append(finalResults, finalResult)
		}
	}

	htmlResult, _ := config.Content.ReadFile("html/results.html")

	// 创建HTML模板
	t := template.Must(template.New("result").Parse(string(htmlResult)))

	// 渲染HTML页面
	if err := t.Execute(w, finalResults); err != nil {
		http.Error(w, "Error rendering HTML page", http.StatusInternalServerError)
		return
	}
}

func mathData(Result config.CheckJsonResult) config.FinalResult {

	var finalResult config.FinalResult
	finalResult.HostIP = Result.HostIP
	checkFiles, _ := ListObjectsInDirectory(checkFileDir)
	for _, file := range checkFiles {
		for v, result := range Result.Results {
			if file.BasicInfo.CheckID == Result.CheckId {
				checkResult := config.CheckResult{
					UID:         result.UID,
					Description: file.Items[v].Description,
					RiskLevel:   file.Items[v].RiskLevel,
					OutPuts:     result.Result.Outputs,
					Harm:        file.Items[v].Harm,
					Solution:    file.Items[v].Solution,
				}

				switch result.Result.Status {
				case 0:
					checkResult.Status = "通过"
				case 2:
					checkResult.Status = "人工检查"
				default:
					checkResult.Status = "失败"
				}
				finalResult.BasicInfo = file.BasicInfo
				finalResult.CheckResults = append(finalResult.CheckResults, checkResult)
			}
		}
	}
	return finalResult
}

func ListObjectsInDirectory(directoryPath string) ([]config.Config, error) {
	var objects []config.Config

	// 获取目录下的所有文件和子目录
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil || len(files) == 0 {
		//utils.Warn(fmt.Sprintf("默认检查配置文件%s不存在，使用内置检查配置文件\n", directoryPath))
		filesList, _ := config.StaticFiles.ReadDir("checkItems")
		for _, file := range filesList {
			fileContent, err := config.StaticFiles.Open("checkItems/" + file.Name())
			if err != nil {
				utils.Warn(fmt.Sprintf("打开%s文件时出错: %v\n", file.Name(), err))
			}
			defer fileContent.Close()
			data, err := ioutil.ReadAll(fileContent)
			if err != nil {
				utils.Warn(fmt.Sprintf("读取%s文件时出错: %v\n", file.Name(), err))
			}
			var tmpResult config.Config

			err = yaml.Unmarshal(data, &tmpResult)
			if err != nil {
				utils.Warn(fmt.Sprintf("解析YAML时出错: %v\n", err))
			}
			// 将对象路径添加到列表中
			objects = append(objects, tmpResult)
		}
	} else {
		// 遍历目录中的每个对象
		for _, file := range files {
			if strings.Contains(file.Name(), "yaml") {
				// 获取对象的完整路径
				objectPath := filepath.Join(directoryPath, file.Name())

				data, err := ioutil.ReadFile(objectPath)
				if err != nil {
					utils.Fatal(fmt.Sprintf("读取%s文件时出错: %v\n", file, err))
				}

				var tmpResult config.Config

				err = yaml.Unmarshal(data, &tmpResult)
				if err != nil {
					utils.Warn(fmt.Sprintf("解析YAML时出错: %v\n", err))
				}
				// 将对象路径添加到列表中
				objects = append(objects, tmpResult)
			}

		}
	}

	if len(objects) == 0 {
		return nil, fmt.Errorf("在目录%s中未找到任何对象", directoryPath)
	} else {
		return objects, nil
	}
}

func generateDeploymentInfoList(checkItemList []config.Config) []map[string]interface{} {
	var deploymentInfoList []map[string]interface{}

	for deploymentId, item := range checkItemList {
		deploymentInfo := make(map[string]interface{})
		switch strings.ToLower(item.BasicInfo.CheckExecutor) {
		case "powershell":
			GeneratePsScript(item)
			deploymentInfo["deploymentCommand"] = fmt.Sprintf(`$url = 'http://%s:8080/%s?key=%s';Invoke-Expression $(Invoke-WebRequest -Uri $url).Content`, ip, item.BasicInfo.CheckID+".ps1", currentAccessKey)
		case "sh":
			GenerateShScript(item)
			deploymentInfo["deploymentCommand"] = fmt.Sprintf("url='http://%s:8080/%s?key=%s';((command -v curl  > /dev/null && curl -sSL $url) || (command -v wget  > /dev/null && wget -qO- $url)) | bash", ip, item.BasicInfo.CheckID+".sh", currentAccessKey)
		default:
			deploymentInfo["deploymentCommand"] = "null"
		}
		deploymentInfo["deploymentId"] = string(deploymentId + 1)
		deploymentInfo["check_id"] = item.BasicInfo.CheckID
		deploymentInfo["check_description"] = item.BasicInfo.CheckDescription
		deploymentInfo["check_name"] = item.BasicInfo.CheckName
		deploymentInfo["last_modified_date"] = item.BasicInfo.LastModifiedDate
		deploymentInfo["check_version"] = item.BasicInfo.CheckVersion
		deploymentInfo["check_type"] = item.BasicInfo.CheckType
		deploymentInfo["operating_system"] = item.BasicInfo.OperatingSystem
		deploymentInfo["additional_information"] = item.BasicInfo.AdditionalInformation
		deploymentInfo["creation_date"] = item.BasicInfo.CreationDate
		deploymentInfo["check_executor"] = item.BasicInfo.CheckExecutor
		deploymentInfoList = append(deploymentInfoList, deploymentInfo)
	}
	return deploymentInfoList
}

func RegenerateAccessKey() {
	// 16 characters when hex encoded
	key := make([]byte, 8)
	rand.Read(key)
	currentAccessKey = hex.EncodeToString(key)
}

func GeneratePsScript(conf config.Config) {

	checkId := conf.BasicInfo.CheckID

	// Generate PowerShell script
	psScript := new(strings.Builder)
	psScript.WriteString("# PowerShell script generated from Go\n\n")
	psScript.WriteString("$OutputEncoding = New-Object -typename System.Text.UTF8Encoding\n\n")
	psScript.WriteString("$results = @()\n\n")

	// Function to get host IP
	psScript.WriteString("function Get-HostIP {\n")
	psScript.WriteString("    return (Test-Connection -ComputerName (hostname) -Count 1).IPv4Address.IPAddressToString\n")
	psScript.WriteString("}\n\n")

	//var finalResults []config.CheckResult
	for _, item := range conf.Items {
		psScript.WriteString(fmt.Sprintf("function %s {\n", item.UID))
		psScript.WriteString(fmt.Sprintf("%s\n", item.Query))
		psScript.WriteString("}\n\n")
		psScript.WriteString(fmt.Sprintf("$result = %s | ConvertFrom-Json\n", item.UID))
		psScript.WriteString(fmt.Sprintf("$results += @{ \"UID\" = \"%s\"; \"result\" = $result }\n\n", item.UID))
	}
	psScript.WriteString("# Construct the final JSON structure\n")
	psScript.WriteString(fmt.Sprintf("$final_result = @{\"check_id\" = \"%s\";\"host.ip\" = (Get-HostIP); \"results\" = $results }\n", checkId))
	psScript.WriteString("$data = $final_result | ConvertTo-Json -Depth 10 -Compress\n")
	psScript.WriteString(fmt.Sprintf("Invoke-RestMethod -Uri \"http://%s:8080/data\" -Method POST -Body $data -ContentType \"application/json; charset=utf-8\"\n", ip))

	fileName := fmt.Sprintf("./config/scripts/%s.ps1", conf.BasicInfo.CheckID)
	err := ioutil.WriteFile(fileName, []byte(psScript.String()), 0644)
	if err != nil {
		utils.Fatal(fmt.Sprintf("保存为%s时出错: %v", fileName, err), color.New(color.FgHiRed).SprintFunc())
	} else {
		utils.Info(fmt.Sprintf("检查项目[%s]脚本已生成,保存到: %s\n", infoColor(conf.BasicInfo.CheckName), fileName))
	}
}

func GenerateShScript(conf config.Config) {

	checkId := conf.BasicInfo.CheckID
	// Generate PowerShell script
	psScript := new(strings.Builder)
	psScript.WriteString(`
get_host_ip() {
echo $(hostname -I | awk '{print $1}')
}
`)

	for _, item := range conf.Items {
		psScript.WriteString(fmt.Sprintf("%s() {\n", item.UID))
		psScript.WriteString(fmt.Sprintf("%s\n", item.Query))
		psScript.WriteString("}\n\n")
		psScript.WriteString(fmt.Sprintf("result=`%s`\n", item.UID))
		psScript.WriteString(fmt.Sprintf("results+=( \"{\\\"UID\\\":\\\"%s\\\",\\\"result\\\": $result}\" )\n\n", item.UID))
	}
	psScript.WriteString("# Construct the final JSON structure\n")
	psScript.WriteString(fmt.Sprintf("data=\"{\\\"check_id\\\": \\\"%s\\\", \\\"host.ip\\\": \\\"$(get_host_ip)\\\", \\\"results\\\": [$(IFS=, ; echo \"${results[*]}\")] }\"\n", checkId))
	psScript.WriteString(fmt.Sprintf("curl -X POST -H \"Content-Type: application/json; charset=utf-8\" -d \"$data\" http://%s:8080/data\n", ip))

	fileName := fmt.Sprintf("./config/scripts/%s.sh", conf.BasicInfo.CheckID)
	err := ioutil.WriteFile(fileName, []byte(psScript.String()), 0644)
	if err != nil {
		utils.Fatal(fmt.Sprintf("保存为[%s]时出错: %v", fileName, err), color.New(color.FgHiRed).SprintFunc())
	} else {
		utils.Info(fmt.Sprintf("检查项目[%s]脚本已生成,保存到: %s\n", infoColor(conf.BasicInfo.CheckName), fileName))
	}

}
