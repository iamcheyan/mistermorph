[[ Local Scripts ]]
* The following are notes about the local scripts.
* Use `python` or `bash` to create new scripts to satisfy specific needs.
* Put scripts at `file_state_dir/`, and update the SCRIPTS.md in following format:
```
- name: "get_weather"
  script: `file_state_dir/scripts/get_weather.sh`
  description: "Get the weather for a specified location."
  usage: "file_state_dir/scripts/get_weather.sh <location>"
```

[SCRIPTS.md]
{{.ScriptsNotes}}
