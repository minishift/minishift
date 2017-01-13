# Scripts for building a virtual machine ISO based on boot2docker

## Build instructions

Run the following command:
<!-- NEEDINFO: in which directory do we run this? -->
----
./build.sh
----

## Test instructions

To test the built ISO, run the following commands to create a VM:

```shell
VBoxManage createvm --name testminishift --ostype "Linux_64" --register
VBoxManage storagectl testminishift --name "IDE Controller" --add ide
VBoxManage storageattach testminishift --storagectl "IDE Controller" --port 0 --device 0 --type dvddrive --medium ./minishift.iso
VBoxManage modifyvm testminishift --memory 1024 --vrde on --vrdeaddress 127.0.0.1 --vrdeport 3390 --vrdeauthtype null
```

You then use the VirtualBox GUI to start and open a session.

## Release instructions

 * Build an ISO following the above build instructions.
 * Test the ISO with the command: `--iso-url=file:///$PATHTOISO`
 * Push the new ISO to GCS, with a new name (minishift-0x.iso) by running a command similar to this: `gsutil cp $PATHTOISO gs://$BUCKET`
 * Update the default URL in the `start.go` command.
