### JSON-schema for Gonkex

Use [file with schema](https://raw.githubusercontent.com/lansfy/gonkex/master/schema/gonkex.json) to add syntax highlight to your favourite IDE and write Gonkex tests more easily.

It adds in-line documentation and auto-completion to any IDE that supports it.

Example in Jetbrains IDE:
![Example Jetbrains](https://i.imgur.com/oYuPuR3.gif)

Example in VSCode IDE:
![Example Jetbrains](https://i.imgur.com/hBIGjP9.gif)

#### Setup in Jetbrains IDE

Download [file with schema](https://raw.githubusercontent.com/lansfy/gonkex/master/gonkex.json).
Open preferences File->Preferences
In Languages & Frameworks > Schemas and DTDs > JSON Schema Mappings

![Jetbrains IDE Settings](https://i.imgur.com/xkO22by.png)

Add new schema

![Add schema](https://i.imgur.com/XHw14GJ.png)

Specify schema name, schema file, and select Schema version: Draft 7

![Name, file, version](https://i.imgur.com/LfJfis0.png)

After that add mapping. You can choose from single file, directory, or file mask.

![Mapping](https://i.imgur.com/iFjm0Ld.png)

Choose what suits you best.

![Mapping pattern](https://i.imgur.com/WIK6sZW.png)

Save your preferences. If you done everything right, you should not see No JSON Schema in bottom right corner

![No Schema](https://i.imgur.com/zLqv1Zv.png)

Instead, you should see your schema name

![Schema Name](https://i.imgur.com/DDXdCO7.png)

#### Setup is VSCode IDE

At first, you need to download YAML Language plugin
Open Extensions by going to Code(File)->Preferences->Extensions

![VSCode Preferences](https://i.imgur.com/X7bk5Kh.png)

Look for YAML and install YAML Language Support by Red Hat

![Yaml Extension](https://i.imgur.com/57onioF.png)

Open Settings by going to Code(File)->Preferences->Settings

Open Schema Settings by typing YAML:Schemas and click on *Edit in settings.json*
![Yaml link](https://i.imgur.com/IEwxWyG.png)

Add file match to apply the JSON on YAML files.

```
"yaml.schemas": {
  "C:\\Users\\Leo\\gonkex.json": ["*.gonkex.yaml"]          
}
```

In the example above the JSON schema stored in C:\Users\Leo\gonkex.json will be applied on all the files that ends with .gonkex.yaml
