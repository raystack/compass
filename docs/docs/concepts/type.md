# Type

Each Asset will have a `Type` to represent the kind of metadata it represents. It is currently pre-defined by Compass, so no arbitrary types will be supported.

Compass currently supports the following types:
1. `table`
2. `job`
3. `dashboard`
4. `topic`
5. `feature_table`
6. `model` (under development)
7. `application` (under development)

Type will be extremely useful for categorizing your assets and it will be really helpful during discovery.
Check [this section on querying assets](../guides/querying#using-the-get-assets-api) on how to leverage `type` for your discovery.
