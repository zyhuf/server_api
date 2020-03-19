Name:           server-api
Version:        6.0.1
Release:        001
Summary:        Reyzar Server API
Vendor :        Reyzar

Group:          System Environment/Daemons
License:        All rights reserved
URL:            http://www.reyzar.net

Source:         server-api.tar.gz

%description
Reyzar Server API.

%prep
%setup -n %{name}

%install
mkdir -p ${RPM_BUILD_ROOT}/opt/reyzar/server-api/
cp -f server-api  ${RPM_BUILD_ROOT}/opt/reyzar/server-api/
chmod +x ${RPM_BUILD_ROOT}/opt/reyzar/server-api/server-api
cp -f config.yaml  ${RPM_BUILD_ROOT}/opt/reyzar/server-api/
mkdir -p ${RPM_BUILD_ROOT}/usr/lib/systemd/system/
cp -f server-api.service ${RPM_BUILD_ROOT}/usr/lib/systemd/system/
chmod +x ${RPM_BUILD_ROOT}/usr/lib/systemd/system/server-api.service

%post
%systemd_post server-api.service

%preun
%systemd_preun server-api.service

%files
/opt/reyzar/server-api/server-api
/opt/reyzar/server-api/config.yaml
/usr/lib/systemd/system/server-api.service