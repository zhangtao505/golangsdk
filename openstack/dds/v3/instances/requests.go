package instances

import (
	"net/http"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/pagination"
)

type CreateOpts struct {
	Name                string          `json:"name"  required:"true"`
	DataStore           DataStore       `json:"datastore" required:"true"`
	Region              string          `json:"region" required:"true"`
	AvailabilityZone    string          `json:"availability_zone" required:"true"`
	VpcId               string          `json:"vpc_id" required:"true"`
	SubnetId            string          `json:"subnet_id" required:"true"`
	SecurityGroupId     string          `json:"security_group_id" required:"true"`
	Password            string          `json:"password,omitempty"`
	Port                string          `json:"port,omitempty"`
	DiskEncryptionId    string          `json:"disk_encryption_id,omitempty"`
	Ssl                 string          `json:"ssl_option,omitempty"`
	Mode                string          `json:"mode" required:"true"`
	Configuration       []Configuration `json:"configurations,omitempty"`
	Flavor              []Flavor        `json:"flavor" required:"true"`
	BackupStrategy      BackupStrategy  `json:"backup_strategy,omitempty"`
	EnterpriseProjectID string          `json:"enterprise_project_id,omitempty"`
	ChargeInfo          *ChargeInfo     `json:"charge_info,omitempty"`
}

type DataStore struct {
	Type          string `json:"type" required:"true"`
	Version       string `json:"version" required:"true"`
	StorageEngine string `json:"storage_engine" required:"true"`
}

type Configuration struct {
	Type string `json:"type" required:"true"`
	Id   string `json:"configuration_id" required:"true"`
}

type Flavor struct {
	Type     string `json:"type" required:"true"`
	Num      int    `json:"num" required:"true"`
	Storage  string `json:"storage,omitempty"`
	Size     int    `json:"size,omitempty"`
	SpecCode string `json:"spec_code" required:"true"`
}

type BackupStrategy struct {
	StartTime string `json:"start_time" required:"true"`
	KeepDays  *int   `json:"keep_days,omitempty"`
	Period    string `json:"period,omitempty"`
}

type BackupPolicyOpts struct {
	BackupPolicy BackupStrategy `json:"backup_policy" required:"true"`
}

type ChargeInfo struct {
	ChargeMode  string `json:"charge_mode" required:"true"`
	PeriodType  string `json:"period_type,omitempty"`
	PeriodNum   int    `json:"period_num,omitempty"`
	IsAutoRenew bool   `json:"is_auto_renew,omitempty"`
	IsAutoPay   bool   `json:"is_auto_pay,omitempty"`
}

type CreateInstanceBuilder interface {
	ToInstancesCreateMap() (map[string]interface{}, error)
}

func (opts CreateOpts) ToInstancesCreateMap() (map[string]interface{}, error) {
	b, err := golangsdk.BuildRequestBody(opts, "")
	if err != nil {
		return nil, err
	}
	return b, nil
}

func Create(client *golangsdk.ServiceClient, opts CreateInstanceBuilder) (r CreateResult) {
	b, err := opts.ToInstancesCreateMap()
	if err != nil {
		r.Err = err
		return
	}

	_, r.Err = client.Post(createURL(client), b, &r.Body, &golangsdk.RequestOpts{
		OkCodes: []int{202},
	})
	return
}

type DeleteInstance struct {
	InstanceId string `json:"instance_id" required:"true"`
}

type DeleteInstanceBuilder interface {
	ToInstancesDeleteMap() (map[string]interface{}, error)
}

func (opts DeleteInstance) ToInstancesDeleteMap() (map[string]interface{}, error) {
	b, err := golangsdk.BuildRequestBody(&opts, "")
	if err != nil {
		return nil, err
	}
	return b, nil
}

func Delete(client *golangsdk.ServiceClient, instanceId string) (r DeleteInstanceResult) {

	url := deleteURL(client, instanceId)

	_, r.Err = client.Delete(url, &golangsdk.RequestOpts{JSONResponse: &r.Body, MoreHeaders: map[string]string{"Content-Type": "application/json"}})
	return
}

type ListInstanceOpts struct {
	Id            string `q:"id"`
	Name          string `q:"name"`
	Mode          string `q:"mode"`
	DataStoreType string `q:"datastore_type"`
	VpcId         string `q:"vpc_id"`
	SubnetId      string `q:"subnet_id"`
	Offset        int    `q:"offset"`
	Limit         int    `q:"limit"`
}

type ListInstanceBuilder interface {
	ToInstanceListDetailQuery() (string, error)
}

func (opts ListInstanceOpts) ToInstanceListDetailQuery() (string, error) {
	q, err := golangsdk.BuildQueryString(opts)
	if err != nil {
		return "", err
	}
	return q.String(), err
}

func List(client *golangsdk.ServiceClient, opts ListInstanceBuilder) pagination.Pager {
	url := listURL(client)
	if opts != nil {
		query, err := opts.ToInstanceListDetailQuery()

		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}

	pageList := pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return InstancePage{pagination.SinglePageBase(r)}
	})

	return pageList
}

// UpdateOpt defines the basic information for update APIs
// URI: <Method> base_url/<Action>
// Request body: {<Param>: <Value>}
// the supported value for Method including: "post" and "put"
type UpdateOpt struct {
	Param  string
	Value  interface{}
	Action string
	Method string
}

type UpdateVolumeOpts struct {
	Volume    VolumeOpts `json:"volume" required:"true"`
	IsAutoPay bool       `json:"is_auto_pay,omitempty"`
}

type VolumeOpts struct {
	GroupID string `json:"group_id,omitempty"`
	Size    *int   `json:"size,omitempty"`
}

type UpdateNodeNumOpts struct {
	Type      string      `json:"type" required:"true"`
	SpecCode  string      `json:"spec_code" required:"true"`
	Num       int         `json:"num" required:"true"`
	Volume    *VolumeOpts `json:"volume,omitempty"`
	IsAutoPay bool        `json:"is_auto_pay,omitempty"`
}

type SpecOpts struct {
	TargetType     string `json:"target_type,omitempty"`
	TargetID       string `json:"target_id" required:"true"`
	TargetSpecCode string `json:"target_spec_code" required:"true"`
}

type UpdateSpecOpts struct {
	Resize    SpecOpts `json:"resize" required:"true"`
	IsAutoPay bool     `json:"is_auto_pay,omitempty"`
}

func Update(client *golangsdk.ServiceClient, instanceId string, opts []UpdateOpt) (r UpdateInstanceResult) {
	for _, optRaw := range opts {
		url := modifyURL(client, instanceId, optRaw.Action)
		var body interface{}
		if optRaw.Param != "" {
			body = map[string]interface{}{
				optRaw.Param: optRaw.Value,
			}
		} else {
			body = optRaw.Value
		}

		var httpMethod func(string, interface{}, interface{}, *golangsdk.RequestOpts) (*http.Response, error)
		if optRaw.Method == "post" {
			httpMethod = client.Post
		} else {
			httpMethod = client.Put
		}

		_, r.Err = httpMethod(url, body, &r.Body, &golangsdk.RequestOpts{
			OkCodes: []int{200, 202},
		})

		if r.Err != nil {
			break
		}
	}
	return
}

var requestOpts golangsdk.RequestOpts = golangsdk.RequestOpts{
	MoreHeaders: map[string]string{"Content-Type": "application/json", "X-Language": "en-us"},
}

// PortOpts is the structure required by the UpdatePort method to modify the database access port.
type PortOpts struct {
	Port int `json:"port"`
}

type EnabledOpts struct {
	Enabled *bool `json:"enabled" required:"true"`
}

type AvailabilityZoneOpts struct {
	TargetAzs string `json:"target_azs" required:"true"`
}

type DescriptionOpts struct {
	Remark string `json:"remark" required:"true"`
}

// UpdateAvailabilityZone is a method to update the AvailabilityZone.
func UpdateAvailabilityZone(c *golangsdk.ServiceClient, instanceId string, opts AvailabilityZoneOpts) (*AvailabilityZoneResp, error) {
	b, err := golangsdk.BuildRequestBody(opts, "")
	if err != nil {
		return nil, err
	}

	var r AvailabilityZoneResp
	_, err = c.Post(availabilityZoneURL(c, instanceId), b, &r, &golangsdk.RequestOpts{
		MoreHeaders: requestOpts.MoreHeaders,
	})
	return &r, err
}

// UpdateRemark is a method to update the description.
func UpdateRemark(c *golangsdk.ServiceClient, instanceId string, opts DescriptionOpts) error {
	b, err := golangsdk.BuildRequestBody(opts, "")
	if err != nil {
		return err
	}

	_, err = c.Put(remarkURL(c, instanceId), b, nil, &golangsdk.RequestOpts{
		MoreHeaders: requestOpts.MoreHeaders,
	})
	return err
}

// UpdateSlowLogStatus is a method to update the slow log status.
func UpdateSlowLogStatus(c *golangsdk.ServiceClient, instanceId string, slowLogStatus string) error {
	_, err := c.Put(updateSlowLogStatusURL(c, instanceId, slowLogStatus), nil, nil, &golangsdk.RequestOpts{
		MoreHeaders: requestOpts.MoreHeaders,
	})
	return err
}

// GetSlowLogStatus is a method to get the slow log status.
func GetSlowLogStatus(c *golangsdk.ServiceClient, instanceId string) (string, error) {
	var r struct {
		Status string `json:"status"`
	}
	_, err := c.Get(getSlowLogStatusURL(c, instanceId), &r, &golangsdk.RequestOpts{
		MoreHeaders: requestOpts.MoreHeaders,
	})
	return r.Status, err
}

// UpdatePort is a method to update the database access port using given parameters.
func UpdatePort(c *golangsdk.ServiceClient, instanceId string, port int) (*PortUpdateResp, error) {
	opts := PortOpts{
		Port: port,
	}
	b, err := golangsdk.BuildRequestBody(opts, "")
	if err != nil {
		return nil, err
	}

	var r PortUpdateResp
	_, err = c.Post(portModifiedURL(c, instanceId), b, &r, &golangsdk.RequestOpts{
		MoreHeaders: requestOpts.MoreHeaders,
	})
	return &r, err
}

// UpdateSecondsLevelMonitoring is a method to update seconds level monitoring.
func UpdateSecondsLevelMonitoring(c *golangsdk.ServiceClient, instanceId string, enabled bool) (*EnabledOpts, error) {
	opts := EnabledOpts{
		Enabled: &enabled,
	}
	b, err := golangsdk.BuildRequestBody(opts, "")
	if err != nil {
		return nil, err
	}

	var r EnabledOpts
	_, err = c.Put(secondsLevelMonitoringURL(c, instanceId), b, &r, &golangsdk.RequestOpts{
		MoreHeaders: requestOpts.MoreHeaders,
		OkCodes: []int{
			204,
		},
	})
	return &r, err
}

// GetSecondsLevelMonitoring is a method to get seconds level monitoring.
func GetSecondsLevelMonitoring(c *golangsdk.ServiceClient, instanceId string) (*EnabledOpts, error) {
	var r EnabledOpts
	_, err := c.Get(secondsLevelMonitoringURL(c, instanceId), &r, &golangsdk.RequestOpts{
		MoreHeaders: requestOpts.MoreHeaders,
	})
	return &r, err
}

// CreateBackupPolicy is a method to create the backup policy.
func CreateBackupPolicy(c *golangsdk.ServiceClient, instanceId string, backPolicy BackupStrategy) (*BackupPolicyResp, error) {
	opts := BackupPolicyOpts{
		BackupPolicy: backPolicy,
	}
	b, err := golangsdk.BuildRequestBody(opts, "")
	if err != nil {
		return nil, err
	}

	var r BackupPolicyResp
	_, err = c.Put(backupPolicyURL(c, instanceId), b, &r, &golangsdk.RequestOpts{
		MoreHeaders: requestOpts.MoreHeaders,
	})
	return &r, err
}

// GetBackupPolicy is a method to get the backup policy.
func GetBackupPolicy(c *golangsdk.ServiceClient, instanceId string) (*BackupPolicyResp, error) {
	var r BackupPolicyResp
	_, err := c.Get(backupPolicyURL(c, instanceId), &r, &golangsdk.RequestOpts{
		MoreHeaders: requestOpts.MoreHeaders,
	})
	return &r, err
}
