package event

import (
	"fmt"
	"os"

	"github.com/helloferdie/stdgo/audittrail/resource"
	"github.com/helloferdie/stdgo/db"
	"github.com/helloferdie/stdgo/eventfaillog"
	"github.com/helloferdie/stdgo/libhttp"
	"github.com/helloferdie/stdgo/librabbitmq"
	"github.com/helloferdie/stdgo/libstring"
)

// CreateAuditTrailAppID -
var CreateAuditTrailAppID = "create-audit-trail"

// CreateAuditTrail -
func CreateAuditTrail(payload map[string]interface{}) {
	mode := os.Getenv("audit_trail_mode")

	if mode == "db" {

	} else {
		r := resource.Get()
		err := librabbitmq.Publish(r, CreateAuditTrailAppID, payload)
		if err != nil {
			// Do failover here
			resp, respCode, err := libhttp.RequestAuditTrails(payload)
			if err != nil || respCode != 200 {
				d, err := db.Open("")
				if err == nil {
					ev := new(eventfaillog.EventFailLog)
					ev.QueueName = r.Queue.Name
					ev.AppID = CreateAuditTrailAppID
					ev.FailoverEndpoint = os.Getenv("audit_trail_url_create")
					ev.Payload = libstring.JSONEncode(payload)
					ev.Remark = fmt.Sprintf("Resp: %v, Error: %v", resp, err)
					ev.Create(d)
				}
			}
		}
	}
}
