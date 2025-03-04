package headscale

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// 全部API响应报文框架
type APIResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type machineData struct {
	Address                []string `json:"addresses"`
	AllowedIPs             []string `json:"allowedIPs"`
	ExtraIPs               []string `json:"extraIPs"`
	AdvertisedIPs          []string `json:"advertisedIPs"`
	HasSubnets             bool     `json:"hasSubnets"`
	AdvertisedExitNode     bool     `json:"advertisedExitNode"`
	AllowedExitNode        bool     `json:"allowedExitNode"`
	HasExitNode            bool     `json:"hasExitNode"` //未实现
	AllowedTags            []string `json:"allowedTags"` //未实现
	InvalidTags            []string `json:"invalidTags"` //未实现
	HasTags                bool     `json:"hasTags"`     //未实现
	Endpoints              []string `json:"endpoints"`
	Derp                   string   `json:"derp"`           //未实现
	IpnVersion             string   `json:"ipnVersion"`     //未实现
	Os                     string   `json:"os"`             //未实现
	Name                   string   `json:"name"`           //未实现
	Fqdn                   string   `json:"fqdn"`           //未实现
	Domain                 string   `json:"domain"`         //未实现
	Created                string   `json:"created"`        //未实现
	Hostname               string   `json:"hostname"`       //未实现
	MachineKey             string   `json:"machineKey"`     //未实现
	NodeKey                string   `json:"nodeKey"`        //未实现
	Id                     string   `json:"id"`             //未实现
	StableId               string   `json:"stableId"`       //未实现
	DisplayNodeKey         string   `json:"displayNodeKey"` //未实现
	LogID                  string   `json:"logID"`          //未实现
	User                   string   `json:"user"`           //未实现
	Creator                string   `json:"creator"`        //未实现
	Expires                string   `json:"expires"`
	NeverExpires           bool     `json:"neverExpires"`
	Authorized             bool     `json:"authorized"`             //未实现
	IsExternal             bool     `json:"isExternal"`             //未实现
	BrokenIPForwarding     bool     `json:"brokenIPForwarding"`     //未实现
	IsEphemeral            bool     `json:"isEphemeral"`            //未实现
	AvailableUpdateVersion string   `json:"availableUpdateVersion"` //未实现
	LastSeen               string   `json:"lastSeen"`               //未实现
	ConnectedToControl     bool     `json:"connectedToControl"`     //未实现
	AutomaticNameMode      bool     `json:"automaticNameMode"`
	TailnetLockKey         string   `json:"tailnetLockKey"` //未实现
}

type machineItem struct {
	Name         string `json:"name"`
	UserAccount  string `json:"useraccount"`
	UserNameHead string `json:"usernamehead"`
	MIPv4        string `json:"mipv4"`
	MIPv6        string `json:"mipv6"`
	OS           string `json:"os"`
	Hostname     string `json:"hostname"`
	Version      string `json:"version"`
	IfOnline     bool   `json:"ifonline"`
	LastSeen     string `json:"lastseen"`
	CreateAt     string `json:"createat"`

	IsSharedIn   bool `json:"issharedin"`
	IsSharedOut  bool `json:"issharedout"`
	NeverExpires bool `json:"neverExpires"`

	AllowedIPs         []string `json:"allowedIPs"`
	ExtraIPs           []string `json:"extraIPs"`
	AdvertisedIPs      []string `json:"advertisedIPs"`
	HasSubnets         bool     `json:"hasSubnets"`
	AdvertisedExitNode bool     `json:"advertisedExitNode"`
	AllowedExitNode    bool     `json:"allowedExitNode"`

	Varies      bool `json:"varies"`
	HairPinning bool `json:"hairpinning"`
	CanIPv6     bool `json:"ipv6en"`
	CanUDP      bool `json:"udpen"`
	CanUPnP     bool `json:"upnpen"`
	CanPCP      bool `json:"pcpen"`
	CanPMP      bool `json:"pmpen"`

	ExpiryDesc string `json:"expirydesc"`

	Endpoints         []string       `json:"endpoints"`
	DERPs             map[string]int `json:"derps"`
	PrefferDERP       string         `json:"usederp"`
	AutomaticNameMode bool           `json:"automaticNameMode"`
}
type adminTemplateConfig struct {
	ErrorMsg     string                 `json:"errormsg"`
	Basedomain   string                 `json:"basedomain"`
	UserName     string                 `json:"username"`
	UserNameHead string                 `json:"usernamehead"`
	UserAccount  string                 `json:"useraccount"`
	OrgName      string                 `json:"orgname"`
	MList        map[string]machineItem `json:"mlist"`
}

// 提供获取用户信息的API
func (h *Headscale) ConsoleSelfAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {
	tokenCookie, _ := req.Cookie("OIDC_Token")
	rawToken := tokenCookie.Value
	idToken, err := h.verifyIDTokenForOIDCCallback(req.Context(), writer, rawToken)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "验证Token失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	claims, err := extractIDTokenClaims(writer, idToken)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "解析用户信息失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	userName, _ /*UID*/, userDisName, err := getUserName(writer, claims, h.cfg.OIDC.StripEmaildomain)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "提取用户信息失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	userNameHead := string([]rune(userDisName)[0])

	userOrgName := userName
	if h.cfg.org_name != "Personal" {
		userOrgName = h.cfg.org_name
	}

	renderData := adminTemplateConfig{
		Basedomain:   h.cfg.BaseDomain,
		UserNameHead: userNameHead,
		UserName:     userDisName,
		UserAccount:  userName,
		OrgName:      userOrgName,
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(&renderData)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

// 验证Token并获取用户信息
func (h *Headscale) verifyTokenIDandGetUser(
	writer http.ResponseWriter,
	req *http.Request,
) string {
	tokenCookie, _ := req.Cookie("OIDC_Token")
	rawToken := tokenCookie.Value
	idToken, err := h.verifyIDTokenForOIDCCallback(req.Context(), writer, rawToken)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "验证Token失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return ""
	}
	claims, err := extractIDTokenClaims(writer, idToken)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "解析用户信息失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return ""
	}
	userName, _ /*UID*/, _ /*userDisName*/, err := getUserName(writer, claims, h.cfg.OIDC.StripEmaildomain)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "提取用户信息失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return ""
	}
	return userName
}

// 控制台获取设备信息列表的API
func (h *Headscale) ConsoleMachinesAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {
	tokenCookie, _ := req.Cookie("OIDC_Token")
	rawToken := tokenCookie.Value
	idToken, err := h.verifyIDTokenForOIDCCallback(req.Context(), writer, rawToken)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "验证Token失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	claims, err := extractIDTokenClaims(writer, idToken)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "解析用户信息失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	userName, _ /*UID*/, _ /*userDisName*/, err := getUserName(writer, claims, h.cfg.OIDC.StripEmaildomain)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "提取用户信息失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}

	UserMachines, err := h.ListMachinesByUser(userName)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "查询用户节点列表失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}

	mlist := make(map[string]machineItem)
	for _, machine := range UserMachines {
		IPNver := machine.HostInfo.IPNVersion
		if strings.Contains(IPNver, "-") {
			IPNver = strings.Split(machine.HostInfo.IPNVersion, "-")[0]
		}
		tz, _ := time.LoadLocation("Asia/Shanghai")

		tmpMachine := machineItem{
			Name:         machine.GivenName,
			UserAccount:  machine.User.Name,
			UserNameHead: string([]rune(machine.User.Display_Name)[0]),
			OS:           machine.HostInfo.OS,
			Hostname:     machine.HostInfo.Hostname,
			Version:      IPNver,
			CreateAt:     machine.CreatedAt.In(tz).Format("2006年01月02日 15:04:05"),
			LastSeen:     machine.LastSeen.In(tz).Format("2006年01月02日 15:04:05"),
			IfOnline:     machine.isOnline(),
			NeverExpires: *machine.Expiry == time.Time{},

			Varies:            machine.HostInfo.NetInfo.MappingVariesByDestIP.EqualBool(true),
			HairPinning:       machine.HostInfo.NetInfo.HairPinning.EqualBool(true),
			CanIPv6:           machine.HostInfo.NetInfo.WorkingIPv6.EqualBool(true),
			CanUDP:            machine.HostInfo.NetInfo.WorkingUDP.EqualBool(true),
			CanUPnP:           machine.HostInfo.NetInfo.UPnP.EqualBool(true),
			CanPCP:            machine.HostInfo.NetInfo.PCP.EqualBool(true),
			CanPMP:            machine.HostInfo.NetInfo.PMP.EqualBool(true),
			Endpoints:         machine.Endpoints,
			AutomaticNameMode: machine.AutoGenName,
		}

		machineRoutes, err := h.GetMachineRoutes(&machine)
		if err != nil {
			errRes := adminTemplateConfig{ErrorMsg: "查询设备路由失败"}
			err = json.NewEncoder(writer).Encode(&errRes)
			if err != nil {
				log.Error().
					Caller().
					Err(err).
					Msg("Failed to write response")
			}
			return
		}
		for _, route := range machineRoutes {
			if route.isExitRoute() {
				if route.Advertised {
					tmpMachine.AdvertisedExitNode = true
					if route.Enabled {
						tmpMachine.AllowedExitNode = true
					}
				}
			} else {
				if route.Advertised {
					tmpMachine.HasSubnets = true
					routeV := netip.Prefix(route.Prefix).String()
					if err != nil {
						errRes := adminTemplateConfig{ErrorMsg: "子网路由地址转换失败"}
						err = json.NewEncoder(writer).Encode(&errRes)
						if err != nil {
							log.Error().
								Caller().
								Err(err).
								Msg("Failed to write response")
						}
						return
					}
					tmpMachine.AdvertisedIPs = append(tmpMachine.AdvertisedIPs, routeV)
					if route.Enabled {
						tmpMachine.AllowedIPs = append(tmpMachine.AllowedIPs, routeV)
					} else {
						tmpMachine.ExtraIPs = append(tmpMachine.ExtraIPs, routeV)
					}
				}
			}
		}

		if machine.HostInfo.NetInfo.PreferredDERP != 0 {
			tmpMachine.DERPs = make(map[string]int)
			for derpname, latency := range machine.HostInfo.NetInfo.DERPLatency {
				ipver := strings.Split(derpname, "-")[1]
				derpname = strings.Split(derpname, "-")[0]
				if ipver == "v4" {
					if peerlatency, ok := machine.HostInfo.NetInfo.DERPLatency[derpname+"-v6"]; ok {
						if latency < peerlatency {
							tmpMachine.DERPs[derpname] = int(latency * 1000)
						}
					} else {
						tmpMachine.DERPs[derpname] = int(latency * 1000)
					}
				} else if ipver == "v6" {
					if peerlatency, ok := machine.HostInfo.NetInfo.DERPLatency[derpname+"-v4"]; ok {
						if latency < peerlatency {
							tmpMachine.DERPs[derpname] = int(latency * 1000)
						}
					} else {
						tmpMachine.DERPs[derpname] = int(latency * 1000)
					}
				} else {
					tmpMachine.DERPs[derpname] = int(latency * 1000)
				}
			}
			tmpMachine.PrefferDERP = strconv.Itoa(machine.HostInfo.NetInfo.PreferredDERP)
		} else {
			tmpMachine.PrefferDERP = "x"
			tmpMachine.DERPs = nil
		}
		if !tmpMachine.NeverExpires {
			ExpiryDuration := machine.Expiry.Sub(time.Now())
			tmpMachine.ExpiryDesc = convExpiryToStr(ExpiryDuration)
		}
		if machine.IPAddresses[0].Is4() {
			tmpMachine.MIPv4 = machine.IPAddresses[0].String()
			tmpMachine.MIPv6 = machine.IPAddresses[1].String()
		} else if machine.IPAddresses[1].Is4() {
			tmpMachine.MIPv6 = machine.IPAddresses[0].String()
			tmpMachine.MIPv4 = machine.IPAddresses[1].String()
		}
		mlist[strconv.FormatUint(machine.ID, 10)] = tmpMachine
	}

	renderData := adminTemplateConfig{
		Basedomain: h.cfg.BaseDomain,
		MList:      mlist,
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(&renderData)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

// 网络设置响应Data体
type NetSettingResData struct {
	FileSharing        bool   `json:"fileSharing"`
	ServicesCollection bool   `json:"servicesCollection"`
	HttpsEnabled       bool   `json:"httpsEnabled"`
	Provider           string `json:"provider"`
	MachineAuthNeeded  bool   `json:"machineAuthNeeded"`
	MaxKeyDurationDays int    `json:"maxKeyDurationDays"`
	NetworkLockEnabled bool   `json:"networkLockEnabled"`
}

// 查询网络设置API
func (h *Headscale) getNetSettingAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {
	userName := h.verifyTokenIDandGetUser(writer, req)
	if userName == "" {
		h.doAPIResponse(writer, "用户信息核对失败", nil)
		return
	}
	user, err := h.GetUser(userName)
	if err != nil {
		h.doAPIResponse(writer, "查询用户失败:"+err.Error(), nil)
		return
	}
	netsettingData := NetSettingResData{
		FileSharing:        false,         //未实现
		ServicesCollection: false,         //未实现
		HttpsEnabled:       false,         //未实现
		Provider:           "Mirage SaaS", //在个人版尚未开启更多验证方式时暂时统一设置
		MachineAuthNeeded:  false,         //未实现
		MaxKeyDurationDays: 180,
		NetworkLockEnabled: false, //未实现
	}
	netsettingData.MaxKeyDurationDays = int(user.ExpiryDuration)
	h.doAPIResponse(writer, "", netsettingData)
}

// 更新用户网络密钥过期时长
func (h *Headscale) ConsoleUpdateKeyExpiryAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {
	userName := h.verifyTokenIDandGetUser(writer, req)
	if userName == "" {
		h.doAPIResponse(writer, "用户信息核对失败", nil)
		return
	}
	err := req.ParseForm()
	if err != nil {
		h.doAPIResponse(writer, "用户请求解析失败:"+err.Error(), nil)
		return
	}
	reqData := make(map[string]int)
	json.NewDecoder(req.Body).Decode(&reqData)
	newExpiryDuration := reqData["maxKeyDurationDays"]
	//	newExpiryDuration, err := strconv.Atoi(newExpiryDurationStr)
	if err != nil {
		h.doAPIResponse(writer, "从请求获取新值失败:"+err.Error(), nil)
		return
	}
	err = h.UpdateUserKeyExpiry(userName, uint(newExpiryDuration))
	if err != nil {
		h.doAPIResponse(writer, "更新密钥过期时长失败:"+err.Error(), nil)
		return
	}
	h.doAPIResponse(writer, "", uint(newExpiryDuration))
}

func (h *Headscale) ConsoleMachinesUpdateAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {
	userName := h.verifyTokenIDandGetUser(writer, req)
	if userName == "" {
		h.doAPIResponse(writer, "用户信息核对失败", nil)
		return
	}
	err := req.ParseForm()
	if err != nil {
		h.doAPIResponse(writer, "用户请求解析失败:"+err.Error(), nil)
		return
	}
	reqData := make(map[string]interface{})
	json.NewDecoder(req.Body).Decode(&reqData)
	reqMID, ok := reqData["mid"].(string)
	if !ok {
		h.doAPIResponse(writer, "用户请求mid解析失败", nil)
		return
	}
	MachineID, err := strconv.ParseUint(reqMID, 0, 64)
	if err != nil {
		h.doAPIResponse(writer, "用户请求mid处理失败", nil)
		return
	}
	toUpdateMachine, err := h.GetMachineByID(MachineID)
	if err != nil {
		h.doAPIResponse(writer, "查询用户设备失败", nil)
		return
	}
	if toUpdateMachine.User.Name != userName {
		h.doAPIResponse(writer, "用户没有该权限", nil)
		return
	}
	reqState, ok := reqData["state"].(string)
	if !ok {
		h.doAPIResponse(writer, "用户请求state解析失败", nil)
		return
	}

	switch reqState {
	case "set-expires": //切换密钥永不过期设置
		msg, err := h.setMachineExpiry(toUpdateMachine)
		if err != nil {
			h.doAPIResponse(writer, msg, nil)
		} else {
			resData := machineData{
				NeverExpires: *toUpdateMachine.Expiry == time.Time{},
				Expires:      msg,
			}
			h.doAPIResponse(writer, "", resData)
		}
	case "rename-node": //设置设备名称
		newName := reqData["nodeName"].(string)
		msg, _, err := h.setMachineName(toUpdateMachine, newName)
		if err != nil {
			h.doAPIResponse(writer, msg, nil)
		} else {
			resData := machineData{
				AutomaticNameMode: toUpdateMachine.AutoGenName,
				Name:              toUpdateMachine.GivenName,
				Hostname:          toUpdateMachine.Hostname,
				NeverExpires:      *toUpdateMachine.Expiry == time.Time{},
				Expires:           msg,
			}
			h.doAPIResponse(writer, "", resData)
		}
	case "set-route-settings": //设置子网转发及出口节点
		allowedIPsInterface := reqData["allowedIPs"].([]interface{})
		allowExitNode := reqData["allowedExitNode"].(bool)

		allowedIPs := new([]string)
		for _, ip := range allowedIPsInterface {
			*allowedIPs = append(*allowedIPs, ip.(string))
		}

		msg, err := h.setMachineSubnet(toUpdateMachine, allowExitNode, *allowedIPs)
		if err != nil {
			h.doAPIResponse(writer, msg, nil)
			return
		} else {
			resData := machineData{
				AutomaticNameMode: toUpdateMachine.AutoGenName,
				Name:              toUpdateMachine.GivenName,
				Hostname:          toUpdateMachine.Hostname,
				NeverExpires:      *toUpdateMachine.Expiry == time.Time{},
				Expires:           msg,
			}
			machineRoutes, err := h.GetMachineRoutes(toUpdateMachine)
			if err != nil {
				h.doAPIResponse(writer, "查询设备路由失败", nil)
				return
			}
			for _, route := range machineRoutes {
				if route.isExitRoute() {
					if route.Advertised {
						resData.AdvertisedExitNode = true
						if route.Enabled {
							resData.AllowedExitNode = true
						}
					}
				} else {
					if route.Advertised {
						resData.HasSubnets = true
						routeV := netip.Prefix(route.Prefix).String()
						if err != nil {
							h.doAPIResponse(writer, "子网路由地址转换失败", nil)
							return
						}
						resData.AdvertisedIPs = append(resData.AdvertisedIPs, routeV)
						if route.Enabled {
							resData.AllowedIPs = append(resData.AllowedIPs, routeV)
						} else {
							resData.ExtraIPs = append(resData.ExtraIPs, routeV)
						}
					}
				}
			}
			h.doAPIResponse(writer, "", resData)
		}
	}
}

// 删除设备API
type removeMachineRes struct {
	Status string `json:"status"`
	ErrMsg string `json:"errmsg"`
}

func (h *Headscale) ConsoleRemoveMachineAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {
	userName := h.verifyTokenIDandGetUser(writer, req)
	resData := removeMachineRes{}
	if userName == "" {
		resData.Status = "Error"
		resData.ErrMsg = "用户信息核对失败"
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		err := json.NewEncoder(writer).Encode(&resData)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	UserMachines, err := h.ListMachinesByUser(userName)
	if err != nil {
		resData.Status = "Error"
		resData.ErrMsg = "用户设备检索失败"
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		err := json.NewEncoder(writer).Encode(&resData)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	err = req.ParseForm()
	if err != nil {
		resData.Status = "Error"
		resData.ErrMsg = "用户请求解析失败"
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		err := json.NewEncoder(writer).Encode(&resData)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	reqData := make(map[string]string)
	json.NewDecoder(req.Body).Decode(&reqData)
	wantRemoveID := reqData["mid"]
	for _, machine := range UserMachines {
		if strconv.FormatUint(machine.ID, 10) == wantRemoveID {
			err = h.HardDeleteMachine(&machine)
			if err != nil {
				resData.Status = "Error"
				resData.ErrMsg = "用户设备删除失败"
				writer.Header().Set("Content-Type", "application/json; charset=utf-8")
				writer.WriteHeader(http.StatusOK)
				err := json.NewEncoder(writer).Encode(&resData)
				if err != nil {
					log.Error().
						Caller().
						Err(err).
						Msg("Failed to write response")
				}
				return
			}
			resData.Status = "OK"
			resData.ErrMsg = "用户设备成功删除"
			writer.Header().Set("Content-Type", "application/json; charset=utf-8")
			writer.WriteHeader(http.StatusOK)
			err := json.NewEncoder(writer).Encode(&resData)
			if err != nil {
				log.Error().
					Caller().
					Err(err).
					Msg("Failed to write response")
			}
			return
		}
	}
	resData.Status = "Error"
	resData.ErrMsg = "未找到目标设备"
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(&resData)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

// API调用的统一响应发报
// @msg 响应状态：成功时data不为nil则忽略，自动设置为success，否则拼接error-{msg}
// @data 响应数据：key值为data的json对象
func (h *Headscale) doAPIResponse(writer http.ResponseWriter, msg string, data interface{}) {
	res := APIResponse{}
	if data != nil {
		res.Status = "success"
		res.Data = data
	} else {
		res.Status = "error-" + msg
	}
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	err := json.NewEncoder(writer).Encode(&res)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

// 切换设备密钥是否禁用过期
func (h *Headscale) setMachineExpiry(machine *Machine) (string, error) {
	if (*machine.Expiry != time.Time{}) {
		err := h.RefreshMachine(machine, time.Time{})
		if err != nil {
			return "设备密钥过期禁用失败", err
		} else {
			return "", err
		}
	} else {
		expiryDuration := time.Hour * 24 * time.Duration(machine.User.ExpiryDuration)
		newExpiry := time.Now().Add(expiryDuration)
		err := h.RefreshMachine(machine, newExpiry)
		if err != nil {
			return "设备密钥过期启用失败", err
		} else {
			return convExpiryToStr(expiryDuration), nil
		}
	}
}

// 三个返回值：msg、nowName、err
func (h *Headscale) setMachineName(machine *Machine, newName string) (string, string, error) {
	newGiveName, err := h.setAutoGenName(machine, newName)
	if err != nil {
		return "设置主机名失败", "", err
	}
	return "", newGiveName, nil
}

func (h *Headscale) setMachineSubnet(machine *Machine, ExitNodeEnable bool, allowedIPs []string) (string, error) {
	machineRoutes, err := h.GetMachineRoutes(machine)
	if err != nil {
		return "获取设备路由设置失败", err
	}
	for _, r := range machineRoutes {
		if r.isExitRoute() {
			if ExitNodeEnable {
				err = h.EnableRoute(uint64(r.ID))
			} else {
				err = h.DisableRoute(uint64(r.ID))
			}
			if err != nil {
				return "设置设备出口节点状态失败", err
			}
		} else {
			err = h.DisableRoute(uint64(r.ID))
		}
	}
	err = h.enableRoutes(machine, allowedIPs...)
	if err != nil {
		return "设置设备子网路由状态失败", err
	}
	return "", nil
}

func convExpiryToStr(duration time.Duration) string {
	if duration.Seconds() <= 0 {
		return "已过期"
	} else if duration.Hours()/24/365 >= 1 {
		return "还剩一年以上有效期"
	} else if duration.Hours()/24/30 >= 1 {
		return "有效期还剩" + strconv.FormatInt(int64(duration.Hours()/24/30), 10) + "个月" + strconv.FormatInt(int64(duration.Hours()/24)-int64(duration.Hours()/24/30)*30, 10) + "天"
	} else if duration.Hours()/24 >= 1 {
		return "有效期还剩" + strconv.FormatInt(int64(duration.Hours()/24), 10) + "天"
	} else if duration.Hours() >= 1 {
		return "有效期还剩" + strconv.FormatInt(int64(duration.Hours()), 10) + "小时"
	} else if duration.Minutes() >= 1 {
		return "有效期还剩" + strconv.FormatInt(int64(duration.Minutes()), 10) + "分钟"
	} else {
		return "马上就要过期"
	}
}

func Time2SHString(t time.Time) string {
	tz, _ := time.LoadLocation("Asia/Shanghai")
	return t.In(tz).Format("2006年01月02日 15:04:05")
}
