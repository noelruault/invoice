# invoice

A simple way to generate invoices.

## Input file example

```yml
fontBody: Arial
fontBodyBold: Arial
fontTitle: Helvetica
fontDescriptorFile: iso-8859-15.map

title: INVOICE
dateFormat: 02/01/2006
footerTitle: Payment details
dateText1: Issue date
dateText2: Due date
```

[input file example](example.yml)

## Config file example

```yml


name: My name
info: |-
  Address to House 123
  ://website | @email | +00 PH0N3 NUMB3R

# Client info
client: |-
  Client address
  Office 1234
  1234, Country
  client@email

# Multiple services can be provided
services:
- description: Software development service
  unitcost: 10
  unit: h.
  quantity: 100
- description: Software QA service
  unitcost: 10
  unit: hours
  quantity: 20

# One or more taxes can be provided
taxes:
- description: VAT
  value: 21
- description: VAT2
  value: 3

currency: €

net: false

# Bank details
account: |-
  Acc Name: My name
  IBAN : XX0000000000000000
  SWIFT: BANKXX00
```

[config file example](config.yml)

## VAT

Value Added Tax is calculated in two different ways, depending on the `net` boolean, which can be set in the `input` template.

- When net is true, the added prices will be reduced by the VAT value(s) provided and added back to the total price at the end of the document.

```
hours: 2
price: 4
net:   true
2 - 2 * 0.21 = 1.58 hour * 4 hours = 8€
```

- If net is false, the prices will be displayed as they are provided, and the VAT value(s) will be added to the total price at the end.

```
hours: 2
price: 4
net:   true
2 hour * 4 hours = 8 + 0.21% = 9,68€
````

## Fonts

```
    fonts/iso-8859-15.map is used to display characters such as '@', that don't come natively with other map encodings.
```

### How to download and use new fonts

- download fonts from [fonts.google.com](https://fonts.google.com)
- place ttf in `fonts/` folder
- run command `make font`, e.g:

```
    echo Belleza-Regular.ttf | make font
    echo Raleway-Regular.ttf | make font
    echo Raleway-SemiBold.ttf | make font
    echo Aldrich-Regular.ttf | make font
```

- change font names (without extension) on config file

```
    fontTitle: Aldrich-Regular
```

## TODO

[ ] Don't override file if same file name is provided
[ ] File name based on ID
