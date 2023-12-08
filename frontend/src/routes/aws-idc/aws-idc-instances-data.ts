import { AwsIdc_ListInstances } from "../../utils/ipc-adapter";

export function awsIdcInstancesData() {
  return AwsIdc_ListInstances()
}