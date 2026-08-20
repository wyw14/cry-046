package domain
import("testing";"time")
func TestDownloadedDeliveryStillExpires(t *testing.T){d:=DeliveryRequest{Status:DeliveryDownloaded,ExpiresAt:time.Unix(10,0)};if d.CanDownload(time.Unix(20,0))==nil{t.Fatal("expired downloaded delivery allowed")}}
