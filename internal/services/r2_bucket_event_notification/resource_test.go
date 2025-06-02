package r2_bucket_event_notification_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/cloudflare/terraform-provider-cloudflare/internal/acctest"
	"github.com/cloudflare/terraform-provider-cloudflare/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func init() {
	resource.AddTestSweepers("cloudflare_r2_bucket_event_notification", &resource.Sweeper{
		Name: "cloudflare_r2_bucket_event_notification",
		F: func(region string) error {
			client, err := acctest.SharedV1Client()
			if err != nil {
				return fmt.Errorf("error establishing client: %w", err)
			}

			accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
			ctx := context.Background()

			// Clean up test queues (which will remove associated event notifications)
			resp, _, err := client.ListQueues(ctx, cloudflare.AccountIdentifier(accountID), cloudflare.ListQueuesParams{})
			if err != nil {
				return err
			}

			for _, q := range resp {
				if strings.HasPrefix(q.Name, "tf-acc-test-") {
					err := client.DeleteQueue(ctx, cloudflare.AccountIdentifier(accountID), q.Name)
					if err != nil {
						return err
					}
				}
			}

			return nil
		},
	})
}

func TestAccCloudflareR2BucketEventNotification_Basic(t *testing.T) {
	rnd := utils.GenerateRandomResourceName()
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	bucketName := "tf-acc-test-" + rnd
	queueName := "tf-acc-test-" + rnd
	resourceName := "cloudflare_r2_bucket_event_notification." + rnd

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.TestAccPreCheck(t)
			acctest.TestAccPreCheck_AccountID(t)
		},
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareR2BucketEventNotificationBasic(rnd, accountID, bucketName, queueName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_id", accountID),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", bucketName),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "jurisdiction", "default"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.0", "PutObject"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.1", "CopyObject"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.description", "Test notification rule"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.prefix", "test/"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.suffix", ".jpg"),
					resource.TestCheckResourceAttrSet(resourceName, "queues.#"),
				),
			},
			{
				Config: testAccCheckCloudflareR2BucketEventNotificationUpdate(rnd, accountID, bucketName, queueName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_id", accountID),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", bucketName),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "jurisdiction", "default"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.0", "PutObject"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.1", "CopyObject"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.2", "DeleteObject"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.description", "Updated notification rule"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.prefix", "updated/"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.suffix", ".png"),
					resource.TestCheckResourceAttr(resourceName, "rules.1.actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.1.actions.0", "CompleteMultipartUpload"),
					resource.TestCheckResourceAttr(resourceName, "rules.1.description", "Second rule"),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return strings.Join([]string{accountID, bucketName, queueName, "default"}, "/"), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCloudflareR2BucketEventNotification_Minimum(t *testing.T) {
	rnd := utils.GenerateRandomResourceName()
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	bucketName := "tf-acc-test-" + rnd
	queueName := "tf-acc-test-" + rnd
	resourceName := "cloudflare_r2_bucket_event_notification." + rnd
	config := testAccCheckCloudflareR2BucketEventNotificationMinimum(rnd, accountID, bucketName, queueName)

	println(config)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.TestAccPreCheck(t)
			acctest.TestAccPreCheck_AccountID(t)
		},
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_id", accountID),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", bucketName),
					resource.TestCheckResourceAttr(resourceName, "jurisdiction", "default"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.0", "PutObject"),
				),
			},
		},
	})
}

func TestAccCloudflareR2BucketEventNotification_Jurisdiction(t *testing.T) {
	rnd := utils.GenerateRandomResourceName()
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	bucketName := "tf-acc-test-" + rnd
	queueName := "tf-acc-test-" + rnd
	resourceName := "cloudflare_r2_bucket_event_notification." + rnd

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.TestAccPreCheck(t)
			acctest.TestAccPreCheck_AccountID(t)
		},
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareR2BucketEventNotificationJurisdiction(rnd, accountID, bucketName, queueName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_id", accountID),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", bucketName),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "jurisdiction", "eu"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actions.0", "PutObject"),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return strings.Join([]string{accountID, bucketName, queueName, "eu"}, "/"), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckCloudflareR2BucketEventNotificationBasic(rnd, accountID, bucketName, queueName string) string {
	return acctest.LoadTestCase("basic.tf", rnd, accountID, bucketName, queueName)
}

func testAccCheckCloudflareR2BucketEventNotificationUpdate(rnd, accountID, bucketName, queueName string) string {
	return acctest.LoadTestCase("update.tf", rnd, accountID, bucketName, queueName)
}

func testAccCheckCloudflareR2BucketEventNotificationMinimum(rnd, accountID, bucketName, queueName string) string {
	return acctest.LoadTestCase("minimum.tf", rnd, accountID, bucketName, queueName)
}

func testAccCheckCloudflareR2BucketEventNotificationJurisdiction(rnd, accountID, bucketName, queueName string) string {
	return acctest.LoadTestCase("jurisdiction.tf", rnd, accountID, bucketName, queueName)
}
