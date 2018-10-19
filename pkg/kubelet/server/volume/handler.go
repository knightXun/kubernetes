package volume

import (
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"net/http"
	"os/exec"
)

const resizeSh = "/usr/local/bin/sync2fs.sh"
const prefix = "3"
const dellType = "dellsc"

type resizeResponse struct {
	// Message for error information.
	// +optional
	Message string `json:"message,omitempty"`

	// code = 200 if ok.
	Code int `json:"code,omitempty"`
}

func CreateHandlers(rootPath string) *restful.WebService {

	ws := &restful.WebService{}
	ws.Path(rootPath).
		Produces(restful.MIME_JSON)

	ws.Route(ws.
		Method("POST").
		Path("/{volumeID}").
		Param(ws.QueryParameter("pretty", "If 'true', then the output is pretty printed.")).
		To(resizeVolume))

	return ws
}

func resizeVolume(request *restful.Request, response *restful.Response) {
	volumeID := request.PathParameter("volumeID")
	volumeType := request.QueryParameter("volumeType")
	glog.V(6).Infof("resize %v volume %v", volumeType, volumeID)
	if volumeType == dellType {
		volumeID = prefix + volumeID
	}
	result := resizeResponse{}
	result.Code = http.StatusBadRequest
	stdoutStderr, err := exec.Command(resizeSh, volumeType, volumeID).CombinedOutput()
	if err != nil {
		result.Message = string(stdoutStderr)
		glog.Errorf("err: %v, stdoutStderr: %v", err, string(stdoutStderr))
	} else {
		result.Code = http.StatusOK
	}
	response.WriteHeaderAndEntity(result.Code, result)
}
