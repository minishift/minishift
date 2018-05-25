Name:           go-hvkvp
Version:        0.0.3
Release:	1%{?dist}
Summary:	Hyper-V Data Exchange binary 

License:        ASL 2.0	
URL:		https://github.com/gbraad/go-hvkvp
Source0:	 %{name}-%{version}.tar.gz

%define BIN_FILE hvkvp

#BuildRequires:	
#Requires:

%description
RPM for installing the hvkvpbinary which is part allows to read key-value pairs used for communication
with the Hyper-V Data Exchange service

%prep
%setup -n %{name}-%{version} -c


%install
rm -rf $RPM_BUILD_ROOT
mkdir -p %{buildroot}/%{_bindir}
cp %{_builddir}/%{name}-%{version}/%{BIN_FILE} %{buildroot}/%{_bindir}/
chmod +x %{buildroot}/%{_bindir}/%{BIN_FILE}


%files
%{_bindir}/%{BIN_FILE}


%changelog
* Wed Sep 05 2017 Gerard Braad <me@gbraad.nl>
- Initial version
