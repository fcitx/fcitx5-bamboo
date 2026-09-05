
file(GLOB BAMBOO_CORE_GO_SRCS bamboo-core/*.go)
file(GLOB BAMBOO_GO_SRCS *.go)
find_program(GO_EXECUTABLE go REQUIRED)

# Custom command for 'go build -buildmode=c-archive ...'
# to create a library from Go codes.
add_custom_command(OUTPUT bamboo-core.a bamboo-core.h
    DEPENDS ${BAMBOO_GO_SRCS} ${BAMBOO_CORE_GO_SRCS} go.mod
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    COMMAND env ${GO_EXECUTABLE} build -buildmode=c-archive
                -trimpath
                -o "${CMAKE_CURRENT_BINARY_DIR}/bamboo-core.a"
            ${BAMBOO_GO_SRCS})
add_custom_target(bamboo-core DEPENDS bamboo-core.a)
# Add a custom target for the library.
add_library(Bamboo::Core UNKNOWN IMPORTED GLOBAL)
add_dependencies(Bamboo::Core bamboo-core)
set_target_properties(Bamboo::Core PROPERTIES
     IMPORTED_LOCATION "${CMAKE_CURRENT_BINARY_DIR}/bamboo-core.a"
     INTERFACE_INCLUDE_DIRECTORIES "${CMAKE_CURRENT_BINARY_DIR}")

if (BUILD_TESTING)
    add_test(NAME bamboo-go-test COMMAND ${GO_EXECUTABLE} test ./...)
    set_tests_properties(bamboo-go-test PROPERTIES
        WORKING_DIRECTORY "${CMAKE_CURRENT_SOURCE_DIR}")
    add_test(NAME bamboo-core-go-test COMMAND ${GO_EXECUTABLE} test ./...)
    set_tests_properties(bamboo-core-go-test PROPERTIES
        WORKING_DIRECTORY "${CMAKE_CURRENT_SOURCE_DIR}/bamboo-core")
endif()
