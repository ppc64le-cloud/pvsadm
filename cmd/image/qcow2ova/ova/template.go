// Copyright 2021 IBM Corp
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ova

var metaTemplate = `os-type = rhel
architecture = ppc64le
vol1-file = {{.ImageName}}
vol1-type = boot`

var ovfTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<ovf:Envelope xmlns:ovf="http://schemas.dmtf.org/ovf/envelope/1" xmlns:rasd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <ovf:References>
    <ovf:File href="{{.VolumeName}}" id="file1" size="{{.SrcVolumeSize}}"/>
  </ovf:References>
  <ovf:DiskSection>
    <ovf:Info>Disk Section</ovf:Info>
    <ovf:Disk capacity="{{.TargetDiskSize}}" capacityAllocationUnits="byte" diskId="disk1" fileRef="file1"/>
  </ovf:DiskSection>
  <ovf:VirtualSystemCollection>
    <ovf:VirtualSystem ovf:id="vs0">
      <ovf:Name>{{.ImageName}}</ovf:Name>
      <ovf:Info></ovf:Info>
      <ovf:ProductSection>
        <ovf:Info/>
        <ovf:Product/>
      </ovf:ProductSection>
      <ovf:OperatingSystemSection ovf:id="79">
        <ovf:Info/>
        <ovf:Description>RHEL</ovf:Description>
        <ns0:architecture xmlns:ns0="ibmpvc">ppc64le</ns0:architecture>
      </ovf:OperatingSystemSection>
      <ovf:VirtualHardwareSection>
        <ovf:Info>Storage resources</ovf:Info>
        <ovf:Item>
          <rasd:Description>Temporary clone for export</rasd:Description>
          <rasd:ElementName>{{.VolumeName}}</rasd:ElementName>
          <rasd:HostResource>ovf:/disk/disk1</rasd:HostResource>
          <rasd:InstanceID>1</rasd:InstanceID>
          <rasd:ResourceType>17</rasd:ResourceType>
          <ns1:boot xmlns:ns1="ibmpvc">True</ns1:boot>
        </ovf:Item>
      </ovf:VirtualHardwareSection>
    </ovf:VirtualSystem>
    <ovf:Info/>
    <ovf:Name>{{.ImageName}}</ovf:Name>
  </ovf:VirtualSystemCollection>
</ovf:Envelope>`
