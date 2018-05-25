go-hvkvp: Hyper-V Data Exchange using Go
========================================

[![pipeline status](https://gitlab.com/gbraad/go-hvkvp/badges/master/pipeline.svg)](https://gitlab.com/gbraad/go-hvkvp/commits/master)


## Prepare and send Key-Value pair
```powershell
$VmMgmt = Get-WmiObject -Namespace root\virtualization\v2 -Class Msvm_VirtualSystemManagementService  
$vm = Get-WmiObject -Namespace root\virtualization\v2 -Class Msvm_ComputerSystem -Filter {ElementName = 'MyVM'}

$kvpDataItem = ([WMIClass][String]::Format("\\{0}\{1}:{2}", $VmMgmt.ClassPath.Server, $VmMgmt.ClassPath.NamespacePath, "Msvm_KvpExchangeDataItem")).CreateInstance()

$kvpDataItem.Name = "IpAddress"
$kvpDataItem.Data = "10.0.75.128"
$kvpDataItem.Source = 0

$VmMgmt.AddKvpItems($Vm, $kvpDataItem.PSBase.GetText(1))
```

#### Note
With `$kvpDataItem.Source = 0` the KVP gets stored as `/var/lib/hyperv/.kvp_pool_0`.


## Receive/Read on the host:
```
$ ./hvkvp
Key: IpAddress, Value: 10.0.75.128
```
