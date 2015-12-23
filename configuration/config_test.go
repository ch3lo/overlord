package configuration

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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

func Test(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

type ConfigSuite struct {
	suite.Suite
	expectedConfig Configuration
}

func (suite *ConfigSuite) SetupTest() {
	os.Clearenv()
	suite.expectedConfig = configStruct
}

func (suite *ConfigSuite) TestMarshalRoundtrip() {
	configBytes, err := yaml.Marshal(suite.expectedConfig)
	assert.Nil(suite.T(), err)
	var config Configuration
	err = yaml.Unmarshal(configBytes, &config)
	assert.Nil(suite.T(), err)
	assert.True(suite.T(), assert.ObjectsAreEqual(config, suite.expectedConfig))
}

func (suite *ConfigSuite) TestParseSimple() {
	var config Configuration
	err := yaml.Unmarshal([]byte(configYaml), &config)
	assert.Nil(suite.T(), err)
	assert.True(suite.T(), assert.ObjectsAreEqual(config, suite.expectedConfig))
}
