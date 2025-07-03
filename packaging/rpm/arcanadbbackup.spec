%global debug_package %{nil}

Name:           arcanadbbackup
Version:        VERSION
Release:        1%{?dist}
Summary:        Arcana DB backup tool

License:        MIT
URL:            https://github.com/nsavinda/arcana-db-backup
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.16
Requires:       postgresql-client

%description
Arcana DB backup tool written in Go for PostgreSQL backups.

%prep
%setup -q -n %{name}-%{version}

%build
# No need for tar extraction here, %setup already handles it
GOOS=linux GOARCH=amd64 go build -o arcanadbbackup .

%install
install -d %{buildroot}%{_bindir}
install -m 755 arcanadbbackup %{buildroot}%{_bindir}/arcanadbbackup

install -d %{buildroot}%{_sysconfdir}/arcanadbbackup
install -m 644 example.config.yaml %{buildroot}%{_sysconfdir}/arcanadbbackup/config.yaml

%files
%license LICENSE
%doc README.md
%{_bindir}/arcanadbbackup
%config(noreplace) %{_sysconfdir}/arcanadbbackup/config.yaml

%changelog
* Wed Jul 03 2025 Nirmal Savinda <nirmalsavinda29@gmail.com> - VERSION-1
- Initial RPM release
