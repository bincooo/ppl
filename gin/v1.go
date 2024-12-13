package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/iocgo/sdk"
	"math/rand"
	"net/http"
	"ppl/core"
	"strconv"
	"strings"
	"time"
)

// @Router()
type Handler struct{}

// @Inject()
func New(container *sdk.Container) *Handler { return &Handler{} }

// @GET(path = "/")
func (h *Handler) index(gtx *gin.Context) {
	gtx.Writer.WriteString("<div style='color:green'>success ~</div>")
}

// @GET(path = "v1/get")
func (h *Handler) get(gtx *gin.Context) {
	var slice []core.Elem
	ty := gtx.DefaultQuery("type", "all")
	re := gtx.DefaultQuery("country", "all")
	so := gtx.DefaultQuery("source", "all")

	count := 0
	for key := range core.ElemMap {
		count++
		elem := core.ElemMap[key]
		if (ty != "all" && elem.T != ty) || (re != "all" && elem.Country != re) || (so != "all" && elem.Source != so) {
			continue
		}
		slice = append(slice, elem)
	}

	size := len(slice)
	if size == 0 {
		gtx.JSON(http.StatusInternalServerError, gin.H{
			"ok":  false,
			"msg": "zero size",
		})
		return
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))
	elem := slice[r.Intn(size)]
	result := fmt.Sprintf("%s://%s:%d", elem.T, elem.Addr, elem.Port)
	gtx.String(http.StatusOK, strings.ToLower(result))
}

// @GET(path = "v1/list")
func (h *Handler) list(gtx *gin.Context) {
	var slice []core.Elem
	ty := gtx.DefaultQuery("type", "all")
	re := gtx.DefaultQuery("country", "all")
	so := gtx.DefaultQuery("source", "all")
	co := gtx.DefaultQuery("count", "1")

	count := 0
	for key := range core.ElemMap {
		count++
		elem := core.ElemMap[key]
		if (ty != "all" && elem.T != ty) || (re != "all" && elem.Country != re) || (so != "all" && elem.Source != so) {
			continue
		}
		slice = append(slice, elem)
	}

	i, err := strconv.Atoi(co)
	if err != nil {
		gtx.JSON(http.StatusOK, gin.H{
			"ok":  false,
			"msg": err.Error(),
		})
		return
	}

	size := len(slice)
	if size < i {
		i = size
	}

	if i == 1 {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		slice = []core.Elem{slice[r.Intn(size)]}
	}

	slice = slice[:i]

	if len(slice) == 0 {
		gtx.JSON(http.StatusOK, gin.H{
			"ok":  false,
			"msg": "zero size",
		})
		return
	}

	gtx.JSON(http.StatusOK, gin.H{
		"ok":   true,
		"data": slice,
		"size": size,
		"_":    count,
	})
}

// @GET(path = "v1/del")
func (h *Handler) del(gtx *gin.Context) {
	addr := gtx.Query("addr")
	port := gtx.Query("port")
	t := gtx.Query("t")
	i, err := strconv.Atoi(port)
	if err != nil {
		gtx.JSON(http.StatusOK, gin.H{
			"ok":  false,
			"msg": err.Error(),
		})
		return
	}

	elem := core.Elem{Addr: addr, Port: i, T: t}
	if _, ok := core.ElemMap[elem.String()]; !ok {
		gtx.JSON(http.StatusOK, gin.H{
			"ok":  false,
			"msg": "not exist: " + elem.String(),
		})
	}

	delete(core.ElemMap, elem.String())
	gtx.JSON(http.StatusOK, gin.H{
		"ok": true,
	})
	return
}
