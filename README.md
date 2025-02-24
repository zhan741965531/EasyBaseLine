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
./EasyBaseLine
```

![图片](https://github.com/user-attachments/assets/7fc6efe4-c272-4703-9c2d-03ee3f14202d)

![图片](https://github.com/user-attachments/assets/5a4e9e4b-43c9-4f12-b997-354ef15a50fb)


2. 远程核查：

```
./EasyBaseLine -remote -host [远程主机地址] -username [用户名] -password [密码] -proto [协议类型] -port [端口]
```
![图片](https://github.com/user-attachments/assets/a7eb89ad-6926-4d21-a1a0-7e4b3abb66c6)


![图片](https://github.com/user-attachments/assets/4718a455-0e99-495b-a1d7-3e4603f200db)

3. 远程批量核查：

```
./EasyBaseLine -remote -file [资产文件路径]
```
![图片](https://github.com/user-attachments/assets/b059025b-6605-4e60-805c-189bca5955f7)

![图片](https://github.com/user-attachments/assets/d9e4ae38-a4be-4c9d-8dbe-6c6462aef6c3)



4. web模式：

```
./EasyBaseLine -server
```

![图片](https://github.com/user-attachments/assets/54384fbf-4534-4653-bec2-009baf60fd17)

![图片](https://github.com/user-attachments/assets/d30a54cc-655c-454d-a4ce-40870c653728)

![图片](https://github.com/user-attachments/assets/7a62a90b-f893-4811-a3dc-b8f256c6593f)

![图片](https://github.com/user-attachments/assets/fe2d4877-93a4-4b59-a9ab-78fbec4cad60)


### 资产文件格式

资产文件应包含一系列资产，每个资产应具有以下格式：

```
Proto(协议类型) Host(主机地址) Username(用户名) Password(密码) Port(端口)
Example:
172.17.15.204 administrator test123 winrm 5985
101.34.19.21 ubuntu test123 ssh 22
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
