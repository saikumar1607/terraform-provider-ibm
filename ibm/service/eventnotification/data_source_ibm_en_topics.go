// Copyright IBM Corp. 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package eventnotification

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	en "github.com/IBM/event-notifications-go-admin-sdk/eventnotificationsv1"
)

func DataSourceIBMEnTopics() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMEnTopicsRead,

		Schema: map[string]*schema.Schema{
			"instance_guid": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique identifier for IBM Cloud Event Notifications instance.",
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of topics.",
			},
			"search_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter the topic by name",
			},
			"topics": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of topics.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Autogenerated topic ID.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the topic.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the topic.",
						},
						"source_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of sources.",
						},
						"sources_names": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of source names.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"subscription_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of subscriptions.",
						},
					},
				},
			},
		},
	}
}

func dataSourceIBMEnTopicsRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	enClient, err := meta.(conns.ClientSession).EventNotificationsApiV1()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, err.Error(), "(Data) ibm_en_topics", "list")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	options := &en.ListTopicsOptions{}

	options.SetInstanceID(d.Get("instance_guid").(string))

	if _, ok := d.GetOk("search_key"); ok {
		options.SetSearch(d.Get("search_key").(string))
	}
	var topicList *en.TopicList

	finalList := []en.TopicsListItem{}

	var offset int64 = 0
	var limit int64 = 100

	options.SetLimit(limit)

	for {
		options.SetOffset(offset)

		result, _, err := enClient.ListTopicsWithContext(context, options)

		topicList = result

		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("ListTopicsWithContext failed: %s", err.Error()), "(Data) ibm_en_topics", "list")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
		offset = offset + limit

		finalList = append(finalList, result.Topics...)

		if offset > *result.TotalCount {
			break
		}
	}

	topicList.Topics = finalList

	d.SetId(fmt.Sprintf("topics_%s", d.Get("instance_guid").(string)))

	if err = d.Set("total_count", flex.IntValue(topicList.TotalCount)); err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Error setting total_count: %s", err), "(Data) ibm_en_topics", "list")
		return tfErr.GetDiag()
	}

	if topicList.Topics != nil {
		err = d.Set("topics", enTopicListFlatten(topicList.Topics))
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Error setting topics: %s", err), "(Data) ibm_en_topics", "list")
			return tfErr.GetDiag()
		}
	}

	return nil
}

func enTopicListFlatten(result []en.TopicsListItem) (topics []map[string]interface{}) {
	for _, topicsItem := range result {
		topics = append(topics, enTopicsToMap(topicsItem))
	}

	return topics
}

func enTopicsToMap(topicsItem en.TopicsListItem) (topicsMap map[string]interface{}) {
	topicsMap = map[string]interface{}{}

	if topicsItem.ID != nil {
		topicsMap["id"] = topicsItem.ID
	}

	if topicsItem.Name != nil {
		topicsMap["name"] = topicsItem.Name
	}

	if topicsItem.Description != nil {
		topicsMap["description"] = topicsItem.Description
	}

	if topicsItem.SourceCount != nil {
		topicsMap["source_count"] = topicsItem.SourceCount
	}

	if topicsItem.SourcesNames != nil {
		topicsMap["sources_names"] = topicsItem.SourcesNames
	}

	if topicsItem.SubscriptionCount != nil {
		topicsMap["subscription_count"] = topicsItem.SubscriptionCount
	}

	return topicsMap
}
