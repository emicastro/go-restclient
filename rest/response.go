package rest

import (
	"container/list"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"
)

// Response structure
type Response struct {
	*http.Response
	Err             error
	byteBody        []byte
	listElement     *list.Element
	skipListElement *skipListNode
	ttl             *time.Time
	lastModified    *time.Time
	etag            string
	revalidate      bool
	cacheHit        atomic.Value
}

func (r *Response) size() int64 {

	size := int64(unsafe.Sizeof(*r))

	size += int64(len(r.byteBody))
	size += int64(unsafe.Sizeof(*r.listElement))
	size += int64(unsafe.Sizeof(*r.skipListElement))
	size += int64(unsafe.Sizeof(*r.ttl))
	size += int64(unsafe.Sizeof(*r.lastModified))
	size += int64(len(r.etag))

	size += int64(len(r.Response.Proto))
	size += int64(len(r.Response.Status))

	return size
}

// String return the Response body as a string
func (r *Response) String() string {
	return string(r.Bytes())
}

// Bytes return the Response body as a slice of bytes.
func (r *Response) Bytes() []byte {
	return r.byteBody
}

// FillUp set the *fill* parameter with the corresponding JSON or XML response.
// fill could be `struct`or `map[string]interface{}`
func (r *Response) FillUp(fill interface{}) error {

	cTypeJSON := "application/json"
	cTypeXML := "application/xml"

	cType := strings.ToLower(r.Header.Get("Content-Type"))

	for i := 0; i < 2; i++ {

		switch {
		case strings.Contains(cType, cTypeJSON):
			return json.Unmarshal(r.byteBody, fill)
		case strings.Contains(cType, cTypeXML):
			return xml.Unmarshal(r.byteBody, fill)
		case i == 0:
			cType = http.DetectContentType(r.byteBody)
		}
	}

	return fmt.Errorf("Response format neither JSON nor XML")
}

// CacheHit shows if a response was get from the cache.
func (r *Response) CacheHit() bool {
	if hit, ok := r.cacheHit.Load().(bool); hit && ok {
		return true
	}
	return false
}

// Debug let any req/res to be dumped, showing how the req/res
// went through the wire, only if debug mode is *on* on RequestBuilder
func (r *Response) Debug() string {
	var strReq, strRes string

	if req, err := httputil.DumpRequest(r.Request, true); err != nil {
		strReq = err.Error()
	} else {
		strReq = string(req)
	}

	if res, err := httputil.DumpResponse(r.Response, true); err != nil {
		strRes = err.Error()
	} else {
		strRes = string(res)
	}

	const separator = "--------\n"
	dump := separator
	dump += "REQUEST\n"
	dump += separator
	dump += strReq
	dump += "\n" + separator
	dump += "RESPONSE\n"
	dump += separator
	dump += strRes
	dump += r.String() + "\n"

	return dump
}
