initial_dir=$(pwd)
for i in $(find "." -type d -name 'pnp*'); do
  # Remove leading './' using parameter expansion
  dir=${i#./}
  module="github.com/go-pnp/go-pnp/${dir}"

  echo "Initializing ${module} in `pwd`/${dir}"
  cd "${dir}"
#  go mod tidy
  cd ${dir} && go mod init ${module} && go mod tidy
  cd "${initial_dir}"
done
