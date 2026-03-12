# kxss
This a adaption of tomnomnom's kxss tool with a different output formt and some optimizations

All Credit for this Code goes to [Tomnomnom](https://github.com/tomnomnom/)
This Tool is under testing and can be unstable sometimes

## Changes to original kxss
I changed the output format of kxss to make it better grepable for my recon script. My new Output Looks like this:
```
URL: https://www.**********.***/event_register.php?event=177 Param: event Unfiltered: [" ' < >]
```

## Installation
To install this Tool please use the following Command:
```
go install github.com/praveenxsharma/kxss@latest
```

## Usage
To run this script use the following command:
```
echo "https://www.**********.***/event_register.php?event=177" | kxss
```

## Question
If you have an question you can create an Issue


