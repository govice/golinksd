# golinksd
[![govice](https://circleci.com/gh/govice/golinksd.svg?style=svg)](https://circleci.com/gh/govice/golinksd)

This is the daemon for the GoVice project. This is currently under development, and you should expect breaking changes. The goal of this project is to produce a blockchain-backed merkle tree used to track the integrity of filesystem(s) over time.

## Usage
```
go install
golinksd
```

## Configuration

See [config.json](/etc/config.json) for an example.

## Docker
```
docker build -t golinksd:latest
docker run -it golinksd -e GOLINKSD_EMAIL=****** -e GOLINKSD_PASSWORD=******
```

## License
Copyright (C) 2019-2020 Kevin Gentile & other contributors (see AUTHORS.md)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.