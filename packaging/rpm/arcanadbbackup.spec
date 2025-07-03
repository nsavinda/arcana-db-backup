Name:           arcanadbbackup
Version:        VERSION
Release:        1%{?dist}
Summary:        Arcana DB backup tool

License:        MIT
URL:            https://github.com/nsavinda/arcana-db-backup
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang

%description
Arcana DB backup tool written in Go for PostgreSQL backups.

%prep
%setup -q

%build
GOOS=linux GOARCH=amd64 go build -o arcanadbbackup main.go

%install
mkdir -p %{buildroot}/usr/local/bin
install -m 755 arcanadbbackup %{buildroot}/usr/local/bin/arcanadbbackup

mkdir -p %{buildroot}/etc/arcanadbbackup
install -m 644 example.config.yaml %{buildroot}/etc/arcanadbbackup/config.yaml

%files
%license LICENSE
%doc README.md
/usr/local/bin/arcanadbbackup
/etc/arcanadbbackup/config.yaml

%changelog
* Wed Jul 03 2025 Nirmal Savinda <nirmalsavinda29@gmail.com> - VERSION-1
- Initial RPM release
