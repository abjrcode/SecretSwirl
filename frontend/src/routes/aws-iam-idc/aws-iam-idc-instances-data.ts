import { ListInstances } from "../../../wailsjs/go/awsiamidc/AwsIdentityCenterController";

export function awsIamIdcInstancesData() {
  return ListInstances()
}