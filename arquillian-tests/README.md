h2d Arquillian Tests
--------------------

Test the [HTTP/2 client (h2d)](https://github.com/rmohid/h2d) against
[Wildfly with HTTP/2 Support in a Docker Image](https://github.com/fstab/docker-wildfly-http2).

The tests run [h2d](https://github.com/rmohid/h2d) as an external command, so [h2d](https://github.com/rmohid/h2d) should be availbale in the `PATH`.

Run with [maven](https://maven.apache.org/) as follows:

```bash
mvn clean package
```

The tests use the [Arquillian Cube Extension](https://github.com/arquillian/arquillian-cube/)
to manage the [Docker](https://www.docker.com) containers.

The current configuration assumes [boot2docker](http://boot2docker.io) listening on port `2376`.
In order to run it on native Linux, remove the property `serverUri` in `src/test/resources/arquillian.xml`.
