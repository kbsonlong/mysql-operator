package k8s

import (
	"context"
	"fmt"

	batchv1 "github.com/kbsonlong/mysql-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetSecretName(name string) (secret_name string) {
	secret_name = fmt.Sprintf("%s-secret", name)
	return secret_name
}

func CretaeOrUpdateSecret(c client.Client, ctx context.Context, ctrl ctrl.Request, cluster *batchv1.MysqlCluster) error {
	secret_name := GetSecretName(cluster.Name)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      secret_name,
		},

		Type: "Opaque",
	}

	err := c.Get(ctx, ctrl.NamespacedName, secret)

	secret.StringData = map[string]string{
		"MYSQL_ROOT_PASSWORD": "123456",
	}
	if err != nil {
		if errors.IsNotFound(err) {
			c.Create(ctx, secret)
		}
		c.Update(ctx, secret)
	}
	return nil
}
