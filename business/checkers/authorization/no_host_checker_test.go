package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
	networking_v1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	security_v1beta "istio.io/client-go/pkg/apis/security/v1beta1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/tests/data"
	"github.com/kiali/kiali/tests/testutils/validations"
)

func TestPresentService(t *testing.T) {
	assert := assert.New(t)

	validations, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"details", "reviews"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "outside"}, models.Namespace{Name: "bookinfo"}},
		ServiceEntries:      map[string][]string{},
		Services:            fakeServices([]string{"details", "reviews"}),
	}.Check()

	// Well configured object
	assert.True(valid)
	assert.Empty(validations)
}

func TestNonExistingService(t *testing.T) {
	assert := assert.New(t)

	vals, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"details", "wrong"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "outside"}, models.Namespace{Name: "bookinfo"}},
		ServiceEntries:      map[string][]string{},
		Services:            fakeServices([]string{"details", "reviews"}),
	}.Check()

	// Wrong host is not present
	assert.False(valid)
	assert.NotEmpty(vals)
	assert.Len(vals, 1)
	assert.Equal(models.ErrorSeverity, vals[0].Severity)
	assert.NoError(validations.ConfirmIstioCheckMessage("authorizationpolicy.nodest.matchingregistry", vals[0]))
	assert.Equal("spec/rules[0]/to[0]/operation/hosts[1]", vals[0].Path)
}

func TestWildcardHost(t *testing.T) {
	assert := assert.New(t)

	vals, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"*", "*.bookinfo", "*.bookinfo.svc.cluster.local"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "outside"}, models.Namespace{Name: "bookinfo"}},
		ServiceEntries:      map[string][]string{},
		Services:            fakeServices([]string{"details", "reviews"}),
	}.Check()

	// Well configured object
	assert.True(valid)
	assert.Empty(vals)
}

func TestWildcardHostOutsideNamespace(t *testing.T) {
	assert := assert.New(t)

	vals, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"*.outside", "*.outside.svc.cluster.local"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "outside"}, models.Namespace{Name: "bookinfo"}},
		ServiceEntries:      map[string][]string{},
		Services:            fakeServices([]string{"details", "reviews"}),
	}.Check()

	assert.False(valid)
	assert.NotEmpty(vals)
	assert.Len(vals, 2)
	assert.Equal(models.ErrorSeverity, vals[0].Severity)
	assert.NoError(validations.ConfirmIstioCheckMessage("authorizationpolicy.nodest.matchingregistry", vals[0]))
	assert.Equal("spec/rules[0]/to[0]/operation/hosts[0]", vals[0].Path)
	assert.Equal(models.ErrorSeverity, vals[1].Severity)
	assert.NoError(validations.ConfirmIstioCheckMessage("authorizationpolicy.nodest.matchingregistry", vals[1]))
	assert.Equal("spec/rules[0]/to[0]/operation/hosts[1]", vals[1].Path)
}

func TestServiceEntryPresent(t *testing.T) {
	assert := assert.New(t)

	serviceEntry := data.CreateExternalServiceEntry()

	validations, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"wikipedia.org"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "outside"}, models.Namespace{Name: "bookinfo"}},
		ServiceEntries:      kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{serviceEntry}),
		Services:            []core_v1.Service{},
	}.Check()

	// Well configured object
	assert.True(valid)
	assert.Empty(validations)
}

func TestExportedInternalServiceEntryPresent(t *testing.T) {
	assert := assert.New(t)

	serviceEntry := data.CreateEmptyMeshInternalServiceEntry("details-se", "bookinfo3", []string{"details.bookinfo2.svc.cluster.local"})

	validations, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"details.bookinfo2.svc.cluster.local"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "bookinfo"}, models.Namespace{Name: "bookinfo2"}, models.Namespace{Name: "bookinfo3"}},
		ServiceEntries:      kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*serviceEntry}),
		Services:            []core_v1.Service{},
	}.Check()

	// Well configured object
	assert.True(valid)
	assert.Empty(validations)
}

func TestExportedExternalServiceEntryPresent(t *testing.T) {
	assert := assert.New(t)

	serviceEntry := data.CreateEmptyMeshExternalServiceEntry("details-se", "bookinfo3", []string{"www.myhost.com"})

	validations, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"www.myhost.com"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "bookinfo"}, models.Namespace{Name: "bookinfo2"}, models.Namespace{Name: "bookinfo3"}},
		ServiceEntries:      kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*serviceEntry}),
		Services:            []core_v1.Service{},
	}.Check()

	// Well configured object
	assert.True(valid)
	assert.Empty(validations)
}

func TestExportedExternalServiceEntryFail(t *testing.T) {
	assert := assert.New(t)

	serviceEntry := data.CreateEmptyMeshExternalServiceEntry("details-se", "bookinfo3", []string{"www.myhost.com"})

	vals, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"www.wrong.com"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "bookinfo"}, models.Namespace{Name: "bookinfo2"}, models.Namespace{Name: "bookinfo3"}},
		ServiceEntries:      kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*serviceEntry}),
		Services:            []core_v1.Service{},
	}.Check()

	// www.wrong.com host is not present
	assert.False(valid)
	assert.NotEmpty(vals)
	assert.Len(vals, 1)
	assert.Equal(models.ErrorSeverity, vals[0].Severity)
	assert.NoError(validations.ConfirmIstioCheckMessage("authorizationpolicy.nodest.matchingregistry", vals[0]))
	assert.Equal("spec/rules[0]/to[0]/operation/hosts[0]", vals[0].Path)
}

func TestWildcardExportedInternalServiceEntryPresent(t *testing.T) {
	assert := assert.New(t)

	serviceEntry := data.CreateEmptyMeshInternalServiceEntry("details-se", "bookinfo3", []string{"*.bookinfo2.svc.cluster.local"})

	validations, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"details.bookinfo2.svc.cluster.local"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "bookinfo"}, models.Namespace{Name: "bookinfo2"}, models.Namespace{Name: "bookinfo3"}},
		ServiceEntries:      kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*serviceEntry}),
		Services:            []core_v1.Service{},
	}.Check()

	// Well configured object
	assert.True(valid)
	assert.Empty(validations)
}

func TestWildcardExportedInternalServiceEntryFail(t *testing.T) {
	assert := assert.New(t)

	serviceEntry := data.CreateEmptyMeshInternalServiceEntry("details-se", "bookinfo3", []string{"details.bookinfo2.svc.cluster.local"})

	vals, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"details.bookinfo3.svc.cluster.local"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "bookinfo"}, models.Namespace{Name: "bookinfo2"}, models.Namespace{Name: "bookinfo3"}},
		ServiceEntries:      kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*serviceEntry}),
		Services:            []core_v1.Service{},
	}.Check()

	// details.bookinfo3.svc.cluster.local host is not present
	assert.False(valid)
	assert.NotEmpty(vals)
	assert.Len(vals, 1)
	assert.Equal(models.ErrorSeverity, vals[0].Severity)
	assert.NoError(validations.ConfirmIstioCheckMessage("authorizationpolicy.nodest.matchingregistry", vals[0]))
	assert.Equal("spec/rules[0]/to[0]/operation/hosts[0]", vals[0].Path)
}

func TestExportedNonFQDNInternalServiceEntryFail(t *testing.T) {
	assert := assert.New(t)

	serviceEntry := data.CreateEmptyMeshInternalServiceEntry("details-se", "bookinfo3", []string{"details"})

	vals, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"details.bookinfo2.svc.cluster.local"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "bookinfo"}, models.Namespace{Name: "bookinfo2"}, models.Namespace{Name: "bookinfo3"}},
		ServiceEntries:      kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*serviceEntry}),
		Services:            []core_v1.Service{},
	}.Check()

	// details.bookinfo2.svc.cluster.local host is not present
	assert.False(valid)
	assert.NotEmpty(vals)
	assert.Len(vals, 1)
	assert.Equal(models.ErrorSeverity, vals[0].Severity)
	assert.NoError(validations.ConfirmIstioCheckMessage("authorizationpolicy.nodest.matchingregistry", vals[0]))
	assert.Equal("spec/rules[0]/to[0]/operation/hosts[0]", vals[0].Path)
}

func TestServiceEntryNotPresent(t *testing.T) {
	assert := assert.New(t)

	serviceEntry := data.CreateExternalServiceEntry()
	vals, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"wrong.org"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "outside"}, models.Namespace{Name: "bookinfo"}},
		ServiceEntries:      kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{serviceEntry}),
		Services:            []core_v1.Service{},
	}.Check()

	// Wrong.org host is not present
	assert.False(valid)
	assert.NotEmpty(vals)
	assert.Len(vals, 1)
	assert.Equal(models.ErrorSeverity, vals[0].Severity)
	assert.NoError(validations.ConfirmIstioCheckMessage("authorizationpolicy.nodest.matchingregistry", vals[0]))
	assert.Equal("spec/rules[0]/to[0]/operation/hosts[0]", vals[0].Path)
}

func TestExportedInternalServiceEntryNotPresent(t *testing.T) {
	assert := assert.New(t)

	serviceEntry := data.CreateEmptyMeshInternalServiceEntry("details-se", "bookinfo3", []string{"details.bookinfo2.svc.cluster.local"})
	vals, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"wrong.bookinfo2.svc.cluster.local"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "bookinfo"}, models.Namespace{Name: "bookinfo2"}, models.Namespace{Name: "bookinfo3"}},
		ServiceEntries:      kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{*serviceEntry}),
		Services:            []core_v1.Service{},
	}.Check()

	// Wrong.org host is not present
	assert.False(valid)
	assert.NotEmpty(vals)
	assert.Len(vals, 1)
	assert.Equal(models.ErrorSeverity, vals[0].Severity)
	assert.NoError(validations.ConfirmIstioCheckMessage("authorizationpolicy.nodest.matchingregistry", vals[0]))
	assert.Equal("spec/rules[0]/to[0]/operation/hosts[0]", vals[0].Path)
}

func TestVirtualServicePresent(t *testing.T) {
	assert := assert.New(t)

	virtualService := *data.CreateEmptyVirtualService("foo-dev", "foo", []string{"foo-dev.example.com"})
	validations, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"foo-dev.example.com"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "outside"}, models.Namespace{Name: "bookinfo"}},
		ServiceEntries:      map[string][]string{},
		Services:            []core_v1.Service{},
		VirtualServices:     []networking_v1alpha3.VirtualService{virtualService},
	}.Check()

	assert.True(valid)
	assert.Empty(validations)
}

func TestVirtualServiceNotPresent(t *testing.T) {
	assert := assert.New(t)

	virtualService := *data.CreateEmptyVirtualService("foo-dev", "foo", []string{"foo-dev.example.com"})
	vals, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"foo-bogus.example.com"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "outside"}, models.Namespace{Name: "bookinfo"}},
		ServiceEntries:      map[string][]string{},
		Services:            []core_v1.Service{},
		VirtualServices:     []networking_v1alpha3.VirtualService{virtualService},
	}.Check()

	// Wrong.org host is not present
	assert.False(valid)
	assert.NotEmpty(vals)
	assert.Len(vals, 1)
	assert.Equal(models.ErrorSeverity, vals[0].Severity)
	assert.NoError(validations.ConfirmIstioCheckMessage("authorizationpolicy.nodest.matchingregistry", vals[0]))
	assert.Equal("spec/rules[0]/to[0]/operation/hosts[0]", vals[0].Path)
}

func TestWildcardServiceEntryHost(t *testing.T) {
	assert := assert.New(t)

	serviceEntry := *data.CreateEmptyMeshExternalServiceEntry("googlecard", "google", []string{"*.google.com"})

	vals, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"maps.google.com"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "outside"}, models.Namespace{Name: "bookinfo"}},
		ServiceEntries:      kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{serviceEntry}),
		Services:            []core_v1.Service{},
	}.Check()

	// Well configured object
	assert.True(valid)
	assert.Empty(vals)

	// Not matching
	vals, valid = NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"maps.apple.com"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "outside"}, models.Namespace{Name: "bookinfo"}},
		ServiceEntries:      kubernetes.ServiceEntryHostnames([]networking_v1alpha3.ServiceEntry{serviceEntry}),
		Services:            []core_v1.Service{},
	}.Check()

	// apple.com host is not present
	assert.False(valid)
	assert.NotEmpty(vals)
	assert.Len(vals, 1)
	assert.Equal(models.ErrorSeverity, vals[0].Severity)
	assert.NoError(validations.ConfirmIstioCheckMessage("authorizationpolicy.nodest.matchingregistry", vals[0]))
	assert.Equal("spec/rules[0]/to[0]/operation/hosts[0]", vals[0].Path)
}

func authPolicyWithHost(hostList []string) *security_v1beta.AuthorizationPolicy {
	methods := []string{"GET", "PUT", "PATCH"}
	nss := []string{"bookinfo"}
	selector := map[string]string{"app": "details", "version": "v1"}
	return data.CreateAuthorizationPolicy(nss, methods, hostList, selector)
}

func fakeServices(serviceNames []string) []core_v1.Service {
	services := make([]core_v1.Service, 0, len(serviceNames))

	for _, sName := range serviceNames {
		service := core_v1.Service{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      sName,
				Namespace: "bookinfo",
				Labels: map[string]string{
					"app":     sName,
					"version": "v1"}},
			Spec: core_v1.ServiceSpec{
				ClusterIP: "fromservice",
				Type:      "ClusterIP",
				Selector:  map[string]string{"app": sName},
			},
		}

		services = append(services, service)
	}

	return services
}

func TestValidServiceRegistry(t *testing.T) {
	assert := assert.New(t)

	validations, valid := NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"ratings.mesh2-bookinfo.svc.mesh1-imports.local"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "outside"}, models.Namespace{Name: "bookinfo"}},
	}.Check()

	assert.False(valid)
	assert.NotEmpty(validations)

	registryService := kubernetes.RegistryStatus{}
	registryService.Hostname = "ratings.mesh2-bookinfo.svc.mesh1-imports.local"

	validations, valid = NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"ratings.mesh2-bookinfo.svc.mesh1-imports.local"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "outside"}, models.Namespace{Name: "bookinfo"}},
		RegistryStatus:      []*kubernetes.RegistryStatus{&registryService},
	}.Check()

	assert.True(valid)
	assert.Empty(validations)

	registryService = kubernetes.RegistryStatus{}
	registryService.Hostname = "ratings2.mesh2-bookinfo.svc.mesh1-imports.local"

	validations, valid = NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"ratings.mesh2-bookinfo.svc.mesh1-imports.local"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "outside"}, models.Namespace{Name: "bookinfo"}},
		RegistryStatus:      []*kubernetes.RegistryStatus{&registryService},
	}.Check()

	assert.False(valid)
	assert.NotEmpty(validations)

	registryService = kubernetes.RegistryStatus{}
	registryService.Hostname = "ratings.bookinfo.svc.cluster.local"

	validations, valid = NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"ratings.bookinfo.svc.cluster.local"}),
		Namespace:           "bookinfo",
		Namespaces:          models.Namespaces{models.Namespace{Name: "outside"}, models.Namespace{Name: "bookinfo"}},
		RegistryStatus:      []*kubernetes.RegistryStatus{&registryService},
	}.Check()

	assert.True(valid)
	assert.Empty(validations)

	registryService = kubernetes.RegistryStatus{}
	registryService.Hostname = "ratings.bookinfo.svc.cluster.local"

	validations, valid = NoHostChecker{
		AuthorizationPolicy: *authPolicyWithHost([]string{"ratings2.bookinfo.svc.cluster.local"}),
		Namespace:           "test",
		Namespaces:          models.Namespaces{models.Namespace{Name: "outside"}, models.Namespace{Name: "bookinfo"}},
		RegistryStatus:      []*kubernetes.RegistryStatus{&registryService},
	}.Check()

	assert.False(valid)
	assert.NotEmpty(validations)
}
