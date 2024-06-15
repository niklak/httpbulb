package httpbulb

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Delete("/delete", MethodsHandle)
	r.Get("/get", MethodsHandle)
	r.Patch("/patch", MethodsHandle)
	r.Post("/post", MethodsHandle)
	r.Put("/put", MethodsHandle)

	r.Get("/headers", http.HandlerFunc(HeadersHandle))
	r.Get("/ip", http.HandlerFunc(IpHandle))
	r.Get("/user-agent", http.HandlerFunc(UserAgentHandle))

	r.Get("/robots.txt", http.HandlerFunc(RobotsHandle))

	r.Get("/gzip", http.HandlerFunc(GzipHandle))
	r.Get("/deflate", http.HandlerFunc(DeflateHandle))
	r.Get("/brotli", http.HandlerFunc(BrotliHandle))
	r.Get("/deny", http.HandlerFunc(DenyHandle))
	r.Get("/encoding/utf8", http.HandlerFunc(Utf8SampleHandle))
	r.Get("/html", http.HandlerFunc(HtmlSampleHandle))
	r.Get("/json", http.HandlerFunc(JSONSampleHandle))
	r.Get("/xml", http.HandlerFunc(XMLSampleHandle))

	r.Handle("/status/{codes}", http.HandlerFunc(StatusCodeHandle))

	r.Handle("/anything", http.HandlerFunc(MethodsHandle))
	r.Handle("/anything/{anything}", http.HandlerFunc(MethodsHandle))

	r.Get("/basic-auth/{user}/{passwd}", http.HandlerFunc(BasicAuthHandle))
	r.Get("/hidden-basic-auth/{user}/{passwd}", http.HandlerFunc(HiddenBasicAuthHandle))
	r.Get("/bearer", http.HandlerFunc(BearerAuthHandle))

	r.Get("/digest-auth/{qop}/{user}/{passwd}", http.HandlerFunc(DigestAuthHandle))
	r.Get("/digest-auth/{qop}/{user}/{passwd}/{algorithm}", http.HandlerFunc(DigestAuthHandle))
	r.Get("/digest-auth/{qop}/{user}/{passwd}/{algorithm}/{stale_after}", http.HandlerFunc(DigestAuthHandle))

	r.Get("/base64/{value}", http.HandlerFunc(Base64DecodeHandle))
	r.Get("/stream/{n:[0-9]+}", http.HandlerFunc(StreamNMessagesHandle))
	r.Get("/stream-bytes/{n:[0-9]+}", http.HandlerFunc(StreamRandomBytesHandle))
	r.Get("/bytes/{n:[0-9]+}", http.HandlerFunc(RandomBytesHandle))
	r.Get("/drip", http.HandlerFunc(DripHandle))
	r.Get("/uuid", http.HandlerFunc(UUIDHandle))
	r.Get("/links/{n:[0-9]+}/{offset:[0-9]+}", http.HandlerFunc(LinkPageHandle))
	r.Get("/links/{n:[0-9]+}", http.HandlerFunc(LinksHandle))
	r.Get("/range/{numbytes:[0-9]+}", http.HandlerFunc(RangeHandle))
	r.Handle("/delay/{delay:[0-9]+}", http.HandlerFunc(DelayHandle))

	r.Get("/cookies", http.HandlerFunc(CookiesHandle))
	r.Get("/cookies-list", http.HandlerFunc(CookiesListHandle))
	r.Get("/cookies/set", http.HandlerFunc(SetCookiesHandle))
	r.Get("/cookies/set/{name}/{value}", http.HandlerFunc(SetCookieHandle))
	r.Get("/cookies/delete", http.HandlerFunc(DeleteCookiesHandle))

	r.Handle("/redirect-to", http.HandlerFunc(RedirectToHandle))
	r.Get("/redirect/{n:[0-9]+}", http.HandlerFunc(RedirectHandle))
	r.Get("/absolute-redirect/{n:[0-9]+}", http.HandlerFunc(AbsoluteRedirectHandle))
	r.Get("/relative-redirect/{n:[0-9]+}", http.HandlerFunc(RelativeRedirectHandle))

	r.Get("/image", http.HandlerFunc(ImageAcceptHandle))
	r.Get("/image/{format:svg|png|jpeg|webp}", http.HandlerFunc(ImageHandle))

	r.Get("/cache", http.HandlerFunc(CacheHandle))
	r.Get("/cache/{value:[0-9]+}", http.HandlerFunc(CacheControlHandle))
	r.Get("/response-headers", http.HandlerFunc(ResponseHeadersHandle))
	r.Post("/response-headers", http.HandlerFunc(ResponseHeadersHandle))
	r.Get("/etag/{etag}", http.HandlerFunc(EtagHandle))
	//TODO: add documentation

	return r
}
