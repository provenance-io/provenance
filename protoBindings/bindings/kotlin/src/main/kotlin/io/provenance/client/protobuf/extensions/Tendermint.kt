package io.provenance.client.protobuf.extensions

import cosmos.base.tendermint.v1beta1.Query
import cosmos.base.tendermint.v1beta1.ServiceGrpc.ServiceBlockingStub as TendermintService

/**
 * Fetches the block at the given height.
 *
 * @param height The height to fetch the requested block at.
 * @return [Query.GetBlockByHeightResponse]
 */
fun TendermintService.getBlockAtHeight(height: Long): Query.GetBlockByHeightResponse =
    getBlockByHeight(Query.GetBlockByHeightRequest.newBuilder().setHeight(height).build())

/**
 * Fetches the current block height.
 *
 * @Return The current block height.
 */
fun TendermintService.getCurrentBlockHeight(): Long =
    getLatestBlock(Query.GetLatestBlockRequest.getDefaultInstance()).block.header.height
