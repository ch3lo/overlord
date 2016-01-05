package api

import (
	"encoding/json"
	"net/http"

	"github.com/ch3lo/overlord/api/types"
	"github.com/ch3lo/overlord/engine"
	"github.com/ch3lo/overlord/manager/service"
	"github.com/ch3lo/overlord/util"
	"github.com/thoas/stats"
	"github.com/unrolled/render"
)

type serviceHandler func(w http.ResponseWriter, r *http.Request) error

type errorHandler struct {
	f serviceHandler
}

func (eh *errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := eh.f(w, r); err != nil {
		util.Log.Errorln(err)
		if err2, ok := err.(apiError); ok {
			jsonRenderer(w, err2)
			return
		}
		jsonRenderer(w, NewUnknownError(err.Error()))
	}
}

type statsHandler struct {
	s *stats.Stats
}

func (sh *statsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stats := sh.s.Data()
	jsonRenderer(w, stats)
}

type Response struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
}

func jsonRenderer(w http.ResponseWriter, i interface{}) {
	rend := render.New()
	rend.JSON(w, http.StatusOK, i)
}

func getServices(w http.ResponseWriter, r *http.Request) error {
	servicesList := engine.GetAppInstance().GetServices()
	var apiServices []types.ServiceGroup
	for _, srv := range servicesList {
		var apiVersions []types.ServiceManager
		for _, v := range srv.Managers {
			var instances []types.Instance
			for _, instance := range v.GetInstances() {
				instances = append(instances, types.Instance{
					Id:           instance.ID,
					CreationDate: &instance.CreationDate,
					Cluster:      instance.ClusterID,
				})
			}

			apiVersions = append(apiVersions, types.ServiceManager{
				Version:      v.Version,
				CreationDate: &srv.CreationDate,
				ImageName:    v.ImageName,
				ImageTag:     v.ImageTag,
				Instances:    instances,
			})
		}
		apiServices = append(apiServices, types.ServiceGroup{
			Id:           srv.ID,
			CreationDate: &srv.CreationDate,
			Managers:     apiVersions,
		})
	}

	jsonRenderer(w, &Response{Status: http.StatusOK, Data: apiServices})
	return nil

}

func putService(w http.ResponseWriter, r *http.Request) error {
	var bindedService types.ServiceGroup
	if err := json.NewDecoder(r.Body).Decode(&bindedService); err != nil {
		return NewSerializationError(err.Error())
	}

	for _, v := range bindedService.Managers {

		clusterCheck := make(map[string]int)
		for k, v := range v.ClusterCheck {
			clusterCheck[k] = v.Instaces
		}

		params := service.Parameters{
			ID:                     bindedService.Id,
			Version:                v.Version,
			ImageName:              v.ImageName,
			ImageTag:               v.ImageTag,
			MinInstancesPerCluster: clusterCheck,
		}

		if _, err := engine.GetAppInstance().RegisterService(params); err != nil {
			switch err.(type) {
			case *service.AlreadyExist:
				return NewElementAlreadyExists()
			case *service.ImageNameRegexpError:
				return NewImageNameRegexpError(err.Error())
			default:
				return NewUnknownError(err.Error())
			}

		}
	}
	jsonRenderer(w, map[string]interface{}{
		"status":  http.StatusOK,
		"service": bindedService})
	return nil
}

func getServiceByServiceId(w http.ResponseWriter, r *http.Request) error {
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

func getServiceByClusterAndServiceId(w http.ResponseWriter, r *http.Request) error {
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

func putServiceVersionByServiceId(w http.ResponseWriter, r *http.Request) error {
	//	serviceId := c.Param("service_id")
	var sv types.ServiceManager

	if err := json.NewDecoder(r.Body).Decode(&sv); err != nil {
		return NewSerializationError(err.Error())
	}
	/*
		managedsv := &manager.ServiceVersion{Version: sv.Version}

		if err := manager.RegisterServiceVersion(serviceId, managedsv); err != nil {
			util.Log.Println(err)
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
