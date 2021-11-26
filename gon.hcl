source    = [".local_dist/sncli_darwin_amd64"]
bundle_id = "uk.co.lessknown.sn-cli"

apple_id {
  username = "jon@lessknown.co.uk"
  password = "@env:AC_PASSWORD"
}

sign {
  application_identity = "Developer ID Application: Jonathan Hadfield (VBZY8FBYR5)"
}