package configmaps

import (
	"fmt"
	"strconv"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateConfigMap(cluster *openldapv1.OpenldapCluster) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.ConfigMapName(),
			Namespace: cluster.Namespace,
			Labels:    cluster.SelectorLabels(),
		},
		Data: defaultConfigMapData(cluster),
	}
}

func defaultConfigMapData(cluster *openldapv1.OpenldapCluster) (data map[string]string) {
	data = map[string]string{
		"LDAP_CUSTOM_LDIF_DIR":       cluster.SeedDataPath(),
		"LDAP_ALLOW_ANON_BINDING":    "no",
		"LDAP_ENABLE_TLS":            utils.ConvertBool(cluster.TlsEnabled()),
		"LDAP_PORT_NUMBER":           strconv.Itoa(int(cluster.LdapPort())),
		"LDAP_ROOT":                  cluster.Spec.OpenldapConfig.Root,
		"LDAP_CONFIG_ADMIN_ENABLED":  "yes",
		"LDAP_ADMIN_USERNAME":        cluster.Spec.OpenldapConfig.AdminUsername,
		"LDAP_CONFIG_ADMIN_USERNAME": cluster.Spec.OpenldapConfig.ConfigUsername,
		"MASTER_HOST": fmt.Sprintf(
			"ldap://%s.%s.svc.cluster.local:%s",
			cluster.WriteServiceName(),
			cluster.Namespace,
			strconv.Itoa(int(cluster.LdapPort())),
		),
	}

	if cluster.TlsEnabled() {
		data["LDAP_LDAPS_PORT_NUMBER"] = strconv.Itoa(int(cluster.LdapsPort()))
		data["LDAP_TLS_CERT_FILE"] = fmt.Sprintf(
			"%s/%s",
			cluster.TlsMountPath(),
			cluster.Spec.OpenldapConfig.Tls.CertFile,
		)
		data["LDAP_TLS_KEY_FILE"] = fmt.Sprintf(
			"%s/%s",
			cluster.TlsMountPath(),
			cluster.Spec.OpenldapConfig.Tls.KeyFile,
		)
		data["LDAP_TLS_CA_FILE"] = fmt.Sprintf(
			"%s/%s",
			cluster.TlsMountPath(),
			cluster.Spec.OpenldapConfig.Tls.CaFile,
		)
	}

	return
}
