package configsvc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/remiges-tech/alya/service"
	"github.com/remiges-tech/alya/wscutils"
	"github.com/remiges-tech/rigel/etcd"
	"github.com/remiges-tech/rigel/server/trees"
	"github.com/remiges-tech/rigel/server/utils"
)

// getSchemaResponse represents the structure for outgoing  responses.

type getConfigResponse struct {
	App         *string  `json:"app,omitempty"`
	Module      *string  `json:"module,omitempty"`
	Version     *int     `json:"ver,omitempty"`
	Config      *string  `json:"config,omitempty"`
	Description string   `json:"description,omitempty"`
	Values      []values `json:"values,omitempty"`
}

type GetConfigRequestParams struct {
	App     *string `form:"app"  binding:"required"`
	Module  *string `form:"module" binding:"required"`
	Version int     `form:"ver" binding:"required"`
	Config  *string `form:"config" binding:"required"`
}

type values struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

func Config_get(c *gin.Context, s *service.Service) {
	lh := s.LogHarbour
	lh.Log("Config_get request received")

	client, ok := s.Dependencies["etcd"].(*etcd.EtcdStorage)
	if !ok {
		field := "etcd"
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, []wscutils.ErrorMessage{wscutils.BuildErrorMessage(utils.INVALID_DEPENDENCY, &field)}))
		return
	}

	var response getConfigResponse
	var queryParams GetConfigRequestParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		lh.LogActivity("Error Unmarshalling Query paramaeters to struct:", err)
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, []wscutils.ErrorMessage{wscutils.BuildErrorMessage(wscutils.ErrcodeInvalidJson, nil)}))
		return
	}

	keyStr := utils.RIGELPREFIX + "/" + *queryParams.App + "/" + *queryParams.Module + "/" + strconv.Itoa(queryParams.Version) + "/fields/" + *queryParams.Config

	getValue, err := client.GetWithPrefix(c, keyStr)
	if err != nil {
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, []wscutils.ErrorMessage{wscutils.BuildErrorMessage(wscutils.ErrcodeMissing, nil, "no_record_found")}))
		lh.Debug0().LogActivity("error while get data from db error:", err.Error)
		return
	} else if getValue != nil {
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, []wscutils.ErrorMessage{wscutils.BuildErrorMessage(wscutils.ErrcodeMissing, nil, "no_record_found")}))
		lh.Debug0().Log("no record found in db")
		return
	}
	// set response fields
	bindGetConfigResponse(&response, &getValue)

	lh.Log(fmt.Sprintf("Record found: %v", map[string]any{"key with --prefix": keyStr, "value": response}))
	wscutils.SendSuccessResponse(c, wscutils.NewSuccessResponse(response))
}

// Config_list: handles the GET /configlist request
func Config_list(c *gin.Context, s *service.Service) {
	lh := s.LogHarbour
	lh.Log("Config_list Request Received")

	// Extracting etcdStorage and rigelTree from service dependency.

	etcd, ok := s.Dependencies["etcd"].(*etcd.EtcdStorage)
	if !ok {
		field := "etcd"
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, []wscutils.ErrorMessage{wscutils.BuildErrorMessage(utils.INVALID_DEPENDENCY, &field)}))
		return
	}
	r := s.Dependencies["rTree"]
	rTree, ok := r.(*utils.Node)
	if !ok {
		field := "rigelTree"
		wscutils.SendErrorResponse(c, wscutils.NewResponse(wscutils.ErrorStatus, nil, []wscutils.ErrorMessage{wscutils.BuildErrorMessage(utils.INVALID_DEPENDENCY, &field)}))
		return
	}

	container := &trees.Container{
		Etcd: etcd,
	}

	trees.Process(rTree, container)

	wscutils.SendSuccessResponse(c, &wscutils.Response{Status: "success", Data: map[string]any{"configurations": container.ResponseData}, Messages: []wscutils.ErrorMessage{}})
}

// bindGetConfigResponse is specifically used in Cinfig_get to bing and set the response
func bindGetConfigResponse(response *getConfigResponse, getValue *map[string]string) {
	for key, vals := range *getValue {

		arry := strings.Split(key, "/")
		keyStr := arry[len(arry)-1]
		if strings.EqualFold(keyStr, "description") {
			response.Description = vals
			ver, _ := strconv.Atoi(arry[5])
			response.App = &arry[3]
			response.Module = &arry[4]
			response.Version = &ver
			response.Config = &arry[7]
			continue
		} else {

			response.Values = append(response.Values, values{
				Name:  keyStr,
				Value: vals,
			})
		}
		ver, _ := strconv.Atoi(arry[5])
		response.App = &arry[3]
		response.Module = &arry[4]
		response.Version = &ver
		response.Config = &arry[7]

	}
}

func getValsForConfigCreateReqError(err validator.FieldError) []string {
	validationErrorVals := GetErrorValidationMapByAPIName("config_create")
	return CommonValidation(validationErrorVals, err)
}

// CommonValidation is a generic function which setup standard validation utilizing
// validator package and Maps the errorVals based on the map parameter and
// return []errorVals
func CommonValidation(validationErrorVals map[string]string, err validator.FieldError) []string {
	var vals []string
	switch err.Tag() {
	case "Required":
		vals = append(vals, "NotProvided")
	}
	return vals
}

func GetErrorValidationMapByAPIName(apiName string) map[string]string {
	var validationsMap = make(map[string]map[string]string)
	validationsMap["config_create"] = map[string]string{
		"Required": "Not_Provided",
	}
	// below is one more example ::

	// validationsMap["country_draft_forward"] = map[string]string{
	// 	"IDmin": "length must be greater than one",
	// }
	return validationsMap[apiName]
}
