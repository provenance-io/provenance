package io.provenance.client.protobuf

import cosmos.base.query.v1beta1.Pagination

/**
 * Create a new pagination request builder given an offset and limit value.
 *
 * @param offset Pagination offset
 * @param limit Pagination limit
 * @return [Pagination.PageRequest.Builder]
 */
fun paginationBuilder(offset: Int, limit: Int): Pagination.PageRequest.Builder =
    Pagination.PageRequest
        .newBuilder()
        .setOffset(offset.toLong())
        .setLimit(limit.toLong())
        .setCountTotal(true)
