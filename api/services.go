package api

import (
	"net/http"

	"github.com/ch3lo/overlord/api/types"
	"github.com/ch3lo/overlord/manager"
	"github.com/ch3lo/overlord/manager/service"
	"github.com/ch3lo/overlord/util"
	"github.com/gin-gonic/gin"
)

func GetServices(c *gin.Context) {
	servicesList := manager.GetAppInstance().GetServices()

	var apiServices []types.Service
	for _, srv := range servicesList {
		var apiVersions []types.ServiceVersion
		for _, v := range srv.Container {
			var instances []types.Instance
			for _, instance := range v.GetInstances() {
				instances = append(instances, types.Instance{
					Id:           instance.Id,
					CreationDate: &instance.CreationDate,
				})
			}

			apiVersions = append(apiVersions, types.ServiceVersion{
				Version:      v.Version,
				CreationDate: &srv.CreationDate,
				ImageName:    v.ImageName,
				ImageTag:     v.ImageTag,
				Instances:    instances,
			})
		}
		apiServices = append(apiServices, types.Service{
			Id:           srv.Id,
			CreationDate: &srv.CreationDate,
			Versions:     apiVersions,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"services": apiServices})
}

func PutService(c *gin.Context) {
	var bindedService types.Service

	if err := c.BindJSON(&bindedService); err != nil {
		util.Log.Println(err)
		se := &SerializationError{Message: err.Error()}
		c.JSON(http.StatusOK, gin.H{
			"status":  se.GetStatus(),
			"message": se.GetMessage()})
		return
	}

	for _, v := range bindedService.Versions {
		params := service.ServiceParameters{
			Id:        bindedService.Id,
			Version:   v.Version,
			ImageName: v.ImageName,
			ImageTag:  v.ImageTag,
		}

		if _, err := manager.GetAppInstance().RegisterService(params); err != nil {
			util.Log.Println(err)

			var newErr CustomStatusAndMessageError
			switch err.(type) {
			case *service.ServiceManagerAlreadyExist:
				newErr = &ElementAlreadyExists{}
				break
			case *service.ImageNameRegexpError:
				newErr = &ImageNameRegexpError{err.Error()}
				break
			default:
				newErr = &UnknownError{}
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  newErr.GetStatus(),
				"message": newErr.GetMessage()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"service": bindedService})
}

func GetServiceByServiceId(c *gin.Context) {
	/*	serviceId := c.Param("service_id")

		bag := manager.ServicesBag()
		for _, v := range bag {
			if v.Id == serviceId {
				c.JSON(http.StatusOK, gin.H{
					"status":  http.StatusOK,
					"service": types.Service{}})
				return
			}
		}*/

	snf := &ServiceNotFound{}
	c.JSON(http.StatusOK, gin.H{
		"status":  snf.GetStatus(),
		"message": snf.GetMessage()})
}

func GetServiceByClusterAndServiceId(c *gin.Context) {
	/*cluster := c.Param("cluster")
	serviceId := c.Param("service_id")

		status, err := manager.GetService(cluster, serviceId)
		if err != nil {
			snf := &ServiceNotFound{}
			c.JSON(http.StatusOK, gin.H{
				"status":  snf.GetStatus(),
				"message": err.Error()})
			return
		}*/

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"service": ""})
}

func PutServiceVersionByServiceId(c *gin.Context) {
	//	serviceId := c.Param("service_id")
	var sv types.ServiceVersion

	if err := c.BindJSON(&sv); err != nil {
		util.Log.Println(err)
		se := &SerializationError{Message: err.Error()}
		c.JSON(http.StatusOK, gin.H{
			"status":  se.GetStatus(),
			"message": se.GetMessage()})
		return
	}
	/*
		managedsv := &manager.ServiceVersion{Version: sv.Version}

		if err := manager.RegisterServiceVersion(serviceId, managedsv); err != nil {
			util.Log.Println(err)
			if ce, ok := err.(*util.ElementAlreadyExists); ok {
				c.JSON(http.StatusOK, gin.H{
					"status":  ce.GetStatus(),
					"message": ce.GetMessage()})
			} else {
				ue := &UnknownError{}
				c.JSON(http.StatusOK, gin.H{
					"status":  ue.GetStatus(),
					"message": ue.GetMessage()})
			}
			return
		}
	*/
	c.JSON(http.StatusOK, gin.H{
		"status":          http.StatusOK,
		"service_version": sv})
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

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"data":   manager.MonitoredServices[0].GetMonitor().Check("asd", "localhost:80")})
}
*/
