module=`go list -m`



projectName=`basename ${module}`
export GOBIN=$projectPath



go install
echo "工程目录:" $projectPath
echo "执行文件:" $projectPath/$projectName
echo 
./$projectName