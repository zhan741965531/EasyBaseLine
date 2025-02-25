# EasyBaseLine

EasyBaseLine 是一个用于执行基线核查的工具，支持本地和远程核查自定义基线。

## 功能

- 本地基线核查
- 远程基线核查
- 远程批量核查，支持凭证文本多资产远程核查
- Server模式，支持一行命令手动登陆核查
- 自动生成核查报告，自动生成excel核查报告

## 使用方法

### 参数说明

```
local           本地核查。
  --checkfile   指定核查文件。
server          Server模式。
  --ip          指定服务IP。
remote          远程核查标识。
  --checkfile   指定核查文件。
  --host        远程主机地址。
  --user        远程主机用户名。
  --password    远程主机密码。
  --proto       协议类型。
  --file        资产文件路径，用于批量核查。
  --port        协议端口。
--help          显示帮助信息。
```

### 示例

1. 本地核查：

```
./EasyBaseLine local
```
![图片](https://github.com/user-attachments/assets/6047bbff-2fa9-4be2-a428-198f380bf8d4)

![图片](https://github.com/user-attachments/assets/9de43d91-e0ba-4d65-b807-4e205b824e15)

2. 远程核查：

```
./EasyBaseLine -remote -host [远程主机地址] -username [用户名] -password [密码] -proto [协议类型] -port [端口]
```
![图片](https://github.com/user-attachments/assets/e9b5b02d-2bde-43ef-a165-15d678830c34)

![图片](https://github.com/user-attachments/assets/1b1002b7-d134-43f3-a0f7-51b0dfe25c21)

3. 远程批量核查：

```
./EasyBaseLine -remote -file [资产文件路径]
```
![图片](https://github.com/user-attachments/assets/ea6613da-04d2-4fd8-b4e5-28246408070b)

![图片](https://github.com/user-attachments/assets/c7172c64-fbab-4338-9101-b294233ecaf9)

4. web模式：

```
./EasyBaseLine -server
```
![图片](https://github.com/user-attachments/assets/d0b20524-a09f-4b5c-bdc5-f28a139be72a)

![图片](https://github.com/user-attachments/assets/0bc4cdb7-c08a-4920-8629-c6e51524a7ce)

![图片](https://github.com/user-attachments/assets/4a0c6471-d870-455b-bf3c-5d0d6644f014)

![图片](https://github.com/user-attachments/assets/ef4c5c20-d2f5-4435-94b4-3066c4e69414)


### 资产文件格式

资产文件应包含一系列资产，每个资产应具有以下格式：

```
Proto(协议类型) Host(主机地址) Username(用户名) Password(密码) Port(端口)
Example:
12.17.15.24 administrator test123 winrm 5985
11.34.19.21 ubuntu test123 ssh 22
```


### 自定义核查文件示例

核查文件是一个固定格式的yaml文件，可以通过以下格式来自行构建核查对象实现自定义核查：

```yaml
basic_info:
  check_id: A0001
  check_name: linux安全基线检查
  check_type: 安全审计
  check_description: |
    Linux安全基线检查，用于评估系统密码策略的合规性。
  check_executor: sh
  operating_system:
    - Ubuntu
    - CentOS
  creation_date: 2023-08-28
  last_modified_date: 2023-08-28
  check_version: 1.0
  additional_information: |
    本检查基于CIS (Center for Internet Security) 的密码策略配置建议进行了定制化。它旨在帮助管理员确保系统密码策略符合最佳实践，以提高系统的安全性。
baseline_check_items:
- uid: CHK001
  description: 检查设备密码复杂度策略
  riskLevel: 高风险
  query: |-
    #!/bin/bash

    declare -A policy_status
    declare -A policy_ref
    files=("/etc/pam.d/system-auth" "/etc/pam.d/passwd" "/etc/pam.d/common-password")

    # ucredit:大写字母个数；lcredit:小写字母个数；dcredit:数字个数；ocredit:特殊字符个数
    policies=("ucredit" "lcredit" "dcredit" "ocredit")

    # 设置参考值
    for policy in "${policies[@]}"
    do
      policy_ref["$policy"]=-1
    done

    # 初始化结果状态为0（合格）
    result_status=0
    outputs="检查设备密码复杂度策略，"

    for file in "${files[@]}"
    do
      if [ -f "$file" ]; then
        for policy in "${policies[@]}"
        do
          # 寻找策略值
          policy_value=$(grep -Po "(?<=${policy}=)-?\d+" "$file" 2>/dev/null)
          if [ -z "$policy_value" ]; then
            outputs+="当前密码复杂度策略不符合要求，找不到 ${policy} 值，"
            result_status=1
          else
            policy_status["$policy"]=$policy_value
            # 检查是否符合期望值，若不符合，将结果状态置为1（不合格）
            if [ "$policy_value" -ne "${policy_ref["$policy"]}" ]; then
              outputs+="当前密码复杂度策略不符合要求，${policy} 的核查值为：${policy_status["$policy"]}，标准值为：${policy_ref["$policy"]}，"
              result_status=1
            fi
          fi
        done
      fi
    done

    # 如果所有的密码策略检查都通过了，添加相应的信息到输出中
    if [ $result_status -eq 0 ]; then
      outputs+="所有的密码复杂度策略都符合要求。"
    fi

    # 根据检查结果，生成json输出
    json_output="{\"outputs\":\"$outputs\",\"status\":$result_status}"
    echo "$json_output"
  expectedOutput: x
  harm: 
    xxx
  solution: |-
    xxx
```

## 作者

爱吃火锅的提莫

## 注意事项

- 请确保有权访问目标主机并执行基线核查，避免非法操作。
- 运行过程中的任何错误都将打印到控制台。

## 反馈与贡献

如果在使用过程中遇到任何问题或有任何建议，欢迎提交 issues 或 pull requests。
