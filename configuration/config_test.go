package configuration

import (
	"os"
	"testing"
	"time"

	"gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

// configStruct is a canonical example configuration, which should map to configYaml
var configStruct = Configuration{
	Updater: Updater{
		Interval: 10 * time.Second,
	},
	Manager: Manager{
		Check: Check{
			Interval:  30 * time.Second,
			Threshold: 4,
		},
	},
	Clusters: map[string]Cluster{
		"dal": Cluster{
			Scheduler: Scheduler{
				"swarm": Parameters{
					"address":   "1.1.1.1:2376",
					"tlsverify": true,
					"tlscacert": "ca-swarm.pem",
					"tlscert":   "cert-swarm.pem",
					"tlskey":    "key-swarm.pem",
				},
			},
		},
		"wdc": Cluster{
			Scheduler: Scheduler{
				"swarm": Parameters{
					"address":   "2.2.2.2:2376",
					"tlsverify": true,
					"tlscacert": "ca-swarm.pem",
					"tlscert":   "cert-swarm.pem",
					"tlskey":    "key-swarm.pem",
				},
			},
		},
		"sjc": Cluster{
			Disabled: true,
			Scheduler: Scheduler{
				"marathon": Parameters{
					"address":   "3.3.3.3:8081",
					"tlsverify": true,
					"tlscacert": "ca-marathon.pem",
					"tlscert":   "cert-marathon.pem",
					"tlskey":    "key-marathon.pem",
				},
			},
		},
	},
	Notification: Notification{
		AttemptsOnError:  5,
		WaitOnError:      10 * time.Second,
		WaitAfterAttemts: 30 * time.Second,
		Providers: map[string]NotificationProvider{
			"email-id": {
				Disabled:         false,
				NotificationType: "email",
				Config: Parameters{
					"from":     "overlord@overlord.com",
					"subject":  "[Notification] bla",
					"smtp":     "smtp.overlord.com",
					"user":     "user",
					"password": "password",
				},
			},
			"email-id2": {
				NotificationType: "email",
				Config: Parameters{
					"from":     "overlord@overlord.com",
					"subject":  "[Notification] bla",
					"smtp":     "smtp.overlord.com",
					"user":     "user",
					"password": "password",
				},
			},
			"rundeck-id": {
				NotificationType: "rundeck",
				Config: Parameters{
					"endpoint": "http://rundeck.com",
					"token":    "qwerty123",
					"job":      "asd321",
				},
			},
		},
	},
}

// configYaml document representing configStruct
var configYaml = `
updater:
  interval: 10s
manager:
  check:
    interval: 30s
    threshold: 4
cluster:
  dal:
    scheduler:
      swarm:
        address: 1.1.1.1:2376
        tlsverify: true
        tlscacert: ca-swarm.pem
        tlscert: cert-swarm.pem
        tlskey: key-swarm.pem
  wdc:
    scheduler:
      swarm:
        address: 2.2.2.2:2376
        tlsverify: true
        tlscacert: ca-swarm.pem
        tlscert: cert-swarm.pem
        tlskey: key-swarm.pem
  sjc:
    disabled: true
    scheduler:
      marathon:
        address: 3.3.3.3:8081
        tlsverify: true
        tlscacert: ca-marathon.pem
        tlscert: cert-marathon.pem
        tlskey: key-marathon.pem
notification:
  attemptsOnError: 5
  waitOnError: 10s
  waitAfterAttemts: 30s 
  providers:
    email-id:
      disabled: false
      type: email
      config:
        from: overlord@overlord.com
        subject: "[Notification] bla"
        smtp: smtp.overlord.com
        user: user
        password: password
    email-id2:
      type: email
      config:
        from: overlord@overlord.com
        subject: "[Notification] bla"
        smtp: smtp.overlord.com
        user: user
        password: password
    rundeck-id:
      type: rundeck
      config:
        endpoint: http://rundeck.com
        token: qwerty123
        job: asd321
`

func Test(t *testing.T) { check.TestingT(t) }

type ConfigSuite struct {
	expectedConfig Configuration
}

var _ = check.Suite(&ConfigSuite{})

func (suite *ConfigSuite) SetUpTest(c *check.C) {
	os.Clearenv()
	suite.expectedConfig = configStruct
}

func (suite *ConfigSuite) TestMarshalRoundtrip(c *check.C) {
	configBytes, err := yaml.Marshal(suite.expectedConfig)
	c.Assert(err, check.IsNil)
	var config Configuration
	err = yaml.Unmarshal(configBytes, &config)
	c.Assert(err, check.IsNil)
	c.Assert(config, check.DeepEquals, suite.expectedConfig)
}

func (suite *ConfigSuite) TestParseSimple(c *check.C) {
	var config Configuration
	err := yaml.Unmarshal([]byte(configYaml), &config)
	c.Assert(err, check.IsNil)
	c.Assert(config, check.DeepEquals, suite.expectedConfig)
}
