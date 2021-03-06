[cmdletbinding()]
param (
    [Parameter(Mandatory)]
    [String]$ApiKey,

    [Parameter(Mandatory)]
    [String]$Path
)

$Headers = @{
    Authorization = "Bearer $ApiKey"
}

$Uri = "https://api.optimizely.com/v2/features"

$Flags = Get-Content -Path $Path -Raw | ConvertFrom-Json

$Features = Invoke-RestMethod -Uri "$($Uri)?project_id=$Env:OPTIMIZELY_PROJECT" -Method GET -Headers $Headers

foreach ($Flag in $Flags.Where({$_.Key -notin $Features.Key})) {

    $Flag | Add-Member -Name project_id -Value ([int64]$Env:OPTIMIZELY_PROJECT) -MemberType NoteProperty

    Invoke-RestMethod -Uri $Uri -Method POST -Body ($Flag | ConvertTo-Json) -Headers $Headers
}
