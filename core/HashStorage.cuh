/*
 * HashStorage.cuh
 *
 *  Created on: Jan 9, 2018
 *      Author: bahadir
 */

#ifndef HASHSTORAGE_CUH_
#define HASHSTORAGE_CUH_

#include <algorithm>
#include <iostream>
#include <numeric>
#include <vector>
#include <fstream>
#include <sstream>
#include <string>
#include <vector>
#include <ctime>
#include "HashStorage.cuh"

#include <thrust/reduce.h>
#include <thrust/device_vector.h>
#include <thrust/host_vector.h>

struct HammingDistanceFilter {
	const uint64_t _target, _maxDistance;

	HammingDistanceFilter(const uint64_t target, const uint64_t maxDistance) :
			_target(target), _maxDistance(maxDistance) {
	}

	__device__ bool operator()(const uint64_t &hash) {
		return __popcll(_target ^ hash) <= _maxDistance;
		//return hammingDistance(_target, hash) <= _maxDistance;
	}
};


class HashStorage {
private:
	uint64_t hashLimit;
	thrust::device_vector<uint64_t> hashes;
public:
	HashStorage(const uint64_t hashLimit);
	virtual ~HashStorage();
	void addHash(uint64_t hash);
	std::vector<uint64_t> search(const uint64_t &hash, const uint64_t maxDistance);

	const thrust::device_vector<uint64_t>* getHashes() const {
		return &hashes;
	}
};

#endif /* HASHSTORAGE_CUH_ */
