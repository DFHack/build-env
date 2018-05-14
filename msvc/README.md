You can [download a free Windows VM from Microsoft](https://developer.microsoft.com/en-us/microsoft-edge/tools/vms/). They're only valid for a few months after the initial startup, but we will be done using it after a few minutes.

Once you have the VM running, download the [Visual C++ 2015 Build Tools installer](https://download.microsoft.com/download/5/F/7/5F7ACAEB-8363-451F-9425-68A90F98B238/visualcppbuildtools_full.exe) and run the following command:

```
.\visualcppbuildtools_full.exe /InstallSelectableItems "NativeLanguageSupport_VCV1;NativeLanguageSupport_XPV1" /Passive /NoRestart
```

This process will take a while, so go get a snack or a drink or play some video games while it runs.

Once it finishes, copy the following directories from the VM to the folder containing this readme.

- C:\\Program Files \(x86\)\\Microsoft Visual Studio 14.0
- C:\\Program Files \(x86\)\\Windows Kits
