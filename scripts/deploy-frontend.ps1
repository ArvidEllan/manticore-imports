param(
    [string]$Stage = "dev",
    [string]$Region = $(if ($env:AWS_REGION) { $env:AWS_REGION } else { "eu-west-1" })
)

$ErrorActionPreference = "Stop"
$ProjectDir = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$WebDir = Join-Path $ProjectDir "web\public"
$BuildDir = Join-Path $ProjectDir ".frontend-build"
$StackName = "manticore-imports-$Stage"

function Get-StackOutput($Key) {
    aws cloudformation describe-stacks `
        --stack-name $StackName `
        --region $Region `
        --query "Stacks[0].Outputs[?OutputKey=='$Key'].OutputValue" `
        --output text
}

$Bucket = Get-StackOutput "FrontendBucketName"
$DistId = Get-StackOutput "CloudFrontDistributionId"
$ApiUrl = Get-StackOutput "ApiUrl"

if (Test-Path $BuildDir) { Remove-Item -Recurse -Force $BuildDir }
New-Item -ItemType Directory -Path $BuildDir | Out-Null
Copy-Item -Path (Join-Path $WebDir "*") -Destination $BuildDir -Recurse

$configJs = @"
window.MANTICORE_CONFIG = {
  apiBaseUrl: "$ApiUrl",
  stage: "$Stage"
};
"@
Set-Content -Path (Join-Path $BuildDir "js\config.js") -Value $configJs -Encoding UTF8

Write-Host "Syncing frontend to s3://$Bucket..."
aws s3 sync $BuildDir "s3://$Bucket" --delete --region $Region --cache-control "public,max-age=3600"

Write-Host "Invalidating CloudFront distribution $DistId..."
aws cloudfront create-invalidation --distribution-id $DistId --paths "/*" --query "Invalidation.Id" --output text

Write-Host "Frontend deployed for stage $Stage"
