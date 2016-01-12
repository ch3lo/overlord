package api

import (
	"encoding/json"
	"net/http"

	valid "github.com/asaskevich/govalidator"
	"github.com/ch3lo/overlord/api/types"
	"github.com/ch3lo/overlord/logger"
	"github.com/ch3lo/overlord/manager/service"
	"github.com/thoas/stats"
	"github.com/unrolled/render"
)

type serviceHandler func(c *appContext, w http.ResponseWriter, r *http.Request) error

type errorHandler struct {
	handler serviceHandler
	appCtx  *appContext
}

func (eh errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := eh.handler(eh.appCtx, w, r); err != nil {
		logger.Instance().Errorln(err)
		if err2, ok := err.(apiError); ok {
			jsonRenderer(w, err2)
			return
		}
		jsonRenderer(w, NewUnknownError(err.Error()))
	}
}

type statsHandler struct {
	*stats.Stats
}

func (sh *statsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	jsonRenderer(w, sh.Data())
}

func jsonRenderer(w http.ResponseWriter, i interface{}) {
	rend := render.New()
	rend.JSON(w, http.StatusOK, i)
}

type Response struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
}

func getServices(c *appContext, w http.ResponseWriter, r *http.Request) error {
	//	servicesList := c.GetApplications()
	var apiServices []types.Application /*
		for _, srv := range servicesList {
			var apiVersions []types.AppVersion
			for _, v := range srv.Managers {
				var instances []types.Instance
				for _, instance := range v.GetInstances() {
					instances = append(instances, types.Instance{
						Id:           instance.ID,
						CreationDate: &instance.CreationDate,
						Cluster:      instance.ClusterID,
					})
				}

				apiVersions = append(apiVersions, types.AppVersion{
					Version:      v.Version,
					CreationDate: &srv.CreationDate,
					ImageName:    v.ImageName,
					ImageTag:     v.ImageTag,
					Instances:    instances,
				})
			}
			apiServices = append(apiServices, types.Application{
				Id:           srv.ID,
				CreationDate: &srv.CreationDate,
				Versions:     apiVersions,
			})
		}*/

	jsonRenderer(w, &Response{Status: http.StatusOK, Data: apiServices})
	return nil

}

func putService(c *appContext, w http.ResponseWriter, r *http.Request) error {
	var appReq types.AppRequest
	if err := json.NewDecoder(r.Body).Decode(&appReq); err != nil {
		return NewSerializationError(err.Error())
	}

	_, err := valid.ValidateStruct(appReq)
	if err != nil {
		return NewSerializationError(err.Error())
	}
	logger.Instance().Debugln("Se valido correctamente la estructura")

	clusterCheck := make(map[string]int)
	for k, v := range appReq.Constraints.ClusterCheck {
		clusterCheck[k] = v.Instances
	}

	params := service.Parameters{
		ID:      appReq.AppID,
		Version: appReq.MajorVersion,
		Constraints: service.ConstraintsParams{
			ImageName:              appReq.Constraints.ImageName,
			MinInstancesPerCluster: clusterCheck,
		},
	}

	if _, err := c.RegisterServiceManager(params); err != nil {
		switch err.(type) {
		case *service.AlreadyExist:
			return NewElementAlreadyExists()
		case *service.ImageNameRegexpError:
			return NewImageNameRegexpError(err.Error())
		default:
			return NewUnknownError(err.Error())
		}
	}

	jsonRenderer(w, map[string]interface{}{
		"status":  http.StatusOK,
		"service": appReq})
	return nil
}

func getServiceByServiceId(c *appContext, w http.ResponseWriter, r *http.Request) error {
	/*	serviceId := c.Param("service_id")

		bag := manager.ServicesBag()
		for _, v := range bag {
			if v.Id == serviceId {
				rend.JSON(http.StatusOK, map[string]string{
					"status":  http.StatusOK,
					"service": types.Service{}})
				return
			}
		}*/
	return NewServiceNotFound()
}

func getServiceByClusterAndServiceId(c *appContext, w http.ResponseWriter, r *http.Request) error {
	/*cluster := c.Param("cluster")
	serviceId := c.Param("service_id")

		status, err := manager.GetService(cluster, serviceId)
		if err != nil {
			snf := &ServiceNotFound{}
			rend.JSON(http.StatusOK, map[string]string{
				"status":  snf.GetStatus(),
				"message": err.Error()})
			return
		}*/
	jsonRenderer(w, map[string]interface{}{
		"status":  http.StatusOK,
		"service": ""})
	return nil
}

func putServiceVersionByServiceId(c *appContext, w http.ResponseWriter, r *http.Request) error {
	//	serviceId := c.Param("service_id")
	var sv types.AppMajorVersion

	if err := json.NewDecoder(r.Body).Decode(&sv); err != nil {
		return NewSerializationError(err.Error())
	}
	/*
		managedsv := &manager.ServiceVersion{Version: sv.Version}

		if err := manager.RegisterServiceVersion(serviceId, managedsv); err != nil {
			logger.Instance().Println(err)
			if ce, ok := err.(*util.ElementAlreadyExists); ok {
				rend.JSON(http.StatusOK, map[string]string{
					"status":  ce.GetStatus(),
					"message": ce.GetMessage()})
			} else {
				ue := &UnknownError{}
				rend.JSON(http.StatusOK, map[string]string{
					"status":  ue.GetStatus(),
					"message": ue.GetMessage()})
			}
			return
		}
	*/
	jsonRenderer(w, map[string]interface{}{
		"status":          http.StatusOK,
		"service_version": sv})
	return nil
}

/*
func ServicesTestGet(c *gin.Context) {

	for i := 0; i < 5; i++ {
		var service manager.Service
		service.Address = fmt.Sprintf("localhost:8%s", i)
		service.Id = strconv.Itoa(i)
		service.Status = "status"
		manager.Register(&service)
	}

	rend.JSON(http.StatusOK, map[string]string{
		"status": http.StatusOK,
		"data":   manager.MonitoredServices[0].GetMonitor().Check("asd", "localhost:80")})
}
*/
