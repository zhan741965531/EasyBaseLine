package config

type Asset struct {
	Host     string `json:"host,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Proto    string `json:"proto,omitempty"`
	Port     int    `json:"port,omitempty"`
}

type Result struct {
	Outputs string `json:"outputs,omitempty"`
	Status  int    `json:"status,omitempty"`
}

type Config struct {
	BasicInfo BasicInfo           `json:"basic_info,omitempty"`
	Items     []BaselineCheckItem `json:"baseline_check_items"`
}

type FinalResult struct {
	HostIP       string        `json:"host.ip,omitempty"`
	BasicInfo    BasicInfo     `json:"basic_info,omitempty"`
	CheckResults []CheckResult `json:"basic_info,omitempty"`
}

type BasicInfo struct {
	CheckID               string   `json:"check_id,omitempty"`
	CheckName             string   `json:"check_name,omitempty"`
	CheckType             string   `json:"check_type,omitempty"`
	CheckDescription      string   `json:"check_description,omitempty"`
	CheckExecutor         string   `json:"check_executor,omitempty"`
	OperatingSystem       []string `json:"operating_system,omitempty"`
	CreationDate          string   `json:"creation_date,omitempty"`
	LastModifiedDate      string   `json:"last_modified_date,omitempty"`
	CheckVersion          float64  `json:"check_version,omitempty"`
	AdditionalInformation string   `json:"additional_information,omitempty"`
}

type BaselineCheckItem struct {
	UID            string `json:"uid,omitempty"`
	Description    string `json:"description,omitempty"`
	RiskLevel      string `json:"riskLevel,omitempty"`
	Query          string `json:"query,omitempty"`
	ExpectedOutput string `json:"expectedOutput,omitempty"`
	Harm           string `json:"harm,omitempty"`
	Solution       string `json:"solution,omitempty"`
}

type CheckResult struct {
	UID         string `json:"UID,omitempty"`
	Description string `json:"description,omitempty"`
	RiskLevel   string `json:"riskLevel,omitempty"`
	Status      string `json:"status,omitempty"`
	OutPuts     string `json:"outPuts,omitempty"`
	Harm        string `json:"harm,omitempty"`
	Solution    string `json:"solution,omitempty"`
}

type CheckJsonResult struct {
	CheckId string `json:"check_id,omitempty"`
	HostIP  string `json:"host.ip,omitempty"`
	Results []struct {
		UID    string `json:"UID,omitempty"`
		Result struct {
			Description string `json:"description,omitempty"`
			RiskLevel   string `json:"riskLevel,omitempty"`
			Harm        string `json:"harm,omitempty"`
			Solution    string `json:"solution,omitempty"`
			Outputs     string `json:"outputs,omitempty"`
			Status      int    `json:"status,omitempty"`
		} `json:"result,omitempty"`
	} `json:"results,omitempty"`
}
