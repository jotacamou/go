package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
)

func main() {
	ec2svc := ec2.New(session.New(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	// Filter out the ec2 instances that are inteded
	// for imaging on each run.  Only those with filter
	// "Backup:true" (in "key:value" form) will be considered
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Backup"),
				Values: []*string{aws.String("true")},
			},
		},
	}
	resp, err := ec2svc.DescribeInstances(params)
	if err != nil {
		klog.Fatal(err)
	}

	// Schedule an imaging request for each of these instances
	for idx, _ := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			var iname *string = inst.InstanceId

			// Overwrite iname if Name tag is set
			for _, tag := range inst.Tags {
				if *tag.Key == "Name" {
					iname = tag.Value
					break
				}
			}

			klog.Infof("Scheduling image creation for %s (%s)", *iname, *inst.InstanceId)

			input := &ec2.CreateImageInput{
				Description: aws.String(fmt.Sprintf("%s image", *iname)),
				InstanceId:  inst.InstanceId,
				Name:        aws.String(fmt.Sprintf("%s-%v", *iname, time.Now().Unix())),
				NoReboot:    aws.Bool(true),
				TagSpecifications: []*ec2.TagSpecification{&ec2.TagSpecification{
					Tags:         inst.Tags,
					ResourceType: aws.String("image"),
				}},
			}

			result, err := ec2svc.CreateImage(input)
			if err != nil {
				klog.Error(err)
				continue
			}

			klog.Infof("Created %s from %s", *result.ImageId, *iname)
		}
	}

	klog.Infof("Rotating AMI images...")

	imagesParams := &ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Backup"),
				Values: []*string{aws.String("true")},
			},
		},
	}

	imagesResp, err := ec2svc.DescribeImages(imagesParams)
	if err != nil {
		klog.Error(err)
	}

	// Group image timestamps per instance
	counts := make(map[string][]string)
	for _, image := range imagesResp.Images {
		// It is expected that the image name will have some
		// arbitrary string and the created timestamp separated
		// by a dash (i.e. myserver-1615610620).  The timestamp
		// is expected to be the last dash separated value.
		elems := strings.Split(*image.Name, "-")
		inst := elems[0]
		ts := elems[len(elems)-1]

		if _, ok := counts[inst]; !ok {
			counts[inst] = make([]string, 0)
		}
		counts[inst] = append(counts[inst], ts+"/"+*image.ImageId)
	}

	// Determine images to rotate (delete)
	rotating := make([]string, 0)
	for k, tsl := range counts {
		// TODO: this should a flag. Sets how many copies will be stored.
		var copies int = 3

		if len(tsl) <= copies {
			klog.Warningf("There's less than two images for %s. Skipping rotation.", k)
			continue
		}

		getTimestamp := func(s string) string {
			return strings.Split(s, "/")[0]
		}

		getId := func(s string) string {
			elems := strings.Split(s, "/")
			return elems[len(elems)-1]
		}

		id, min := getId(tsl[0]), getTimestamp(tsl[0])
		//id := getId(tsl[0])
		for _, ts := range tsl {
			if getTimestamp(ts) < min {
				min = getTimestamp(ts)
				id = getId(ts)
			}
		}
		rotating = append(rotating, id)
	}

	// Do the rotation
	for _, image := range rotating {
		deregisterInput := &ec2.DeregisterImageInput{
			ImageId: &image,
		}
		_, err := ec2svc.DeregisterImage(deregisterInput)
		if err != nil {
			klog.Warningf("Could not deregister %s: %v", image, err)
		}
		klog.Infof("Removed %s", image)
	}

}
