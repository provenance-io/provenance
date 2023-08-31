package io.provenance.client.protobuf.extensions

import cosmos.base.tendermint.v1beta1.Query
import cosmos.base.tendermint.v1beta1.ServiceGrpc.ServiceBlockingStub as BlockingTendermintService
import cosmos.base.tendermint.v1beta1.ServiceGrpcKt.ServiceCoroutineStub as CoroutineTendermintService

/**
 * Fetches the block at the given height.
 *
 * @param height The height to fetch the requested block at.
 * @return [Query.GetBlockByHeightResponse]
 */
fun BlockingTendermintService.getBlockAtHeight(height: Long): Query.GetBlockByHeightResponse =
    getBlockByHeight(Query.GetBlockByHeightRequest.newBuilder().setHeight(height).build())

/**
 * Fetches the block at the given height.
 *
 * @param height The height to fetch the requested block at.
 * @return [Query.GetBlockByHeightResponse]
 */
suspend fun CoroutineTendermintService.getBlockAtHeight(height: Long): Query.GetBlockByHeightResponse =
    getBlockByHeight(Query.GetBlockByHeightRequest.newBuilder().setHeight(height).build())

/**
 * Fetches the current block height.
 *
 * @Return The current block height.
 */
fun BlockingTendermintService.getCurrentBlockHeight(): Long =
    getLatestBlock(Query.GetLatestBlockRequest.getDefaultInstance()).block.header.height

/**
 * Fetches the current block height.
 *
 * @Return The current block height.
 */
suspend fun CoroutineTendermintService.getCurrentBlockHeight(): Long =
    getLatestBlock(Query.GetLatestBlockRequest.getDefaultInstance()).block.header.height
