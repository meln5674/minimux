package minimux_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/meln5674/minimux"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Simple", func() {
	It("should wrap a plain Handler", func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		})

		req, err := http.NewRequest(http.MethodGet, "http://localhost/", nil)
		Expect(err).ToNot(HaveOccurred())
		resp1 := httptest.NewRecorder()
		mux.ServeHTTP(resp1, req)

		req, err = http.NewRequest(http.MethodGet, "http://localhost/", nil)
		Expect(err).ToNot(HaveOccurred())
		resp2 := httptest.NewRecorder()
		Expect(minimux.Simple(mux).ServeHTTP(context.Background(), resp2, req, nil, nil)).ToNot(HaveOccurred())

		Expect(resp1).To(Equal(resp2))
	})
})

var _ = Describe("SimpleFunc", func() {
	It("should wrap a plain HandlerFunc", func() {
		f := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		})

		req, err := http.NewRequest(http.MethodGet, "http://localhost/", nil)
		Expect(err).ToNot(HaveOccurred())
		resp1 := httptest.NewRecorder()
		f.ServeHTTP(resp1, req)

		req, err = http.NewRequest(http.MethodGet, "http://localhost/", nil)
		Expect(err).ToNot(HaveOccurred())
		resp2 := httptest.NewRecorder()
		Expect(minimux.SimpleFunc(f).ServeHTTP(context.Background(), resp2, req, nil, nil)).ToNot(HaveOccurred())

		Expect(resp1).To(Equal(resp2))
	})
})

var _ = Describe("NotFound", func() {
	It("should return 404 and no body", func() {
		req, err := http.NewRequest(http.MethodGet, "http://localhost/", nil)
		Expect(err).ToNot(HaveOccurred())
		resp := httptest.NewRecorder()
		Expect(minimux.NotFound.ServeHTTP(context.Background(), resp, req, nil, nil)).To(Succeed())
		Expect(resp.Code).To(Equal(http.StatusNotFound))
		Expect(resp.Body.String()).To(BeEmpty())
	})
})

var _ = Describe("Redirecting", func() {
	It("should perform the redirect", func() {
		req, err := http.NewRequest(http.MethodGet, "http://localhost/", nil)
		Expect(err).ToNot(HaveOccurred())
		resp := httptest.NewRecorder()
		Expect(minimux.RedirectingTo("/foo", http.StatusFound).ServeHTTP(context.Background(), resp, req, nil, nil)).To(Succeed())
		Expect(resp.Code).To(Equal(http.StatusFound))
		location, err := resp.Result().Location()
		Expect(err).ToNot(HaveOccurred())
		Expect(location.String()).To(Equal("/foo"))
	})
})
