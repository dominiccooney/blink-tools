# See <https://gist.github.com/233296> for how to convert from page_cycler CSV
# in format:
#
# When,Site,Min,Max,Mean,StdDev,T1,T2,T3,T4,T5,T6,T7,T8,T9,T10,...
# Before,www.example.com,101,109,102.89,0.87,...
# Before,www.wikipedia.org,100,120,108.44,4.74,...
#
# To a long data frame in this format:
# [1] "Site"     "When"     "variable" "value"

library('reshape')

ReadPageCyclerResults <- function(filename) {
  # Load a CSV file of results from page_cycler.
  #
  # To generate this format with a manual run of page cycler:
  #  1. Run the page_cycler
  #  2. On the results page: Select all. Copy.
  #  3. Paste it into Google Sheets in Column B.
  #  4. Put a repeating label (for example, a revision number) in Column A.
  #     Multiple results can be accumulated this way in one spreadsheet.
  #  5. File, Download As, CSV.
  #  6. Use grep label-from-Column-A to filter out the summary data (such
  #     as timer lag, etc.)
  #
  # Args:
  #   filename: The name of the CSV file to load.
  #
  # Returns:
  #   A long frame with these columns:
  #   Site: A factor with the identifier of the page_cycler test.
  #   When: A factor with the label from Column A.
  #   variable: A factor with the iteration of the test, T1, T2, ...
  #   value: The result of the iteration.

  csv <- read.csv(filename, header=FALSE)
  num.cols <- dim(csv)[2]
  fixed.col.names <- c('When', 'Site', 'Min', 'Max', 'Mean', 'Std.d')
  num.results <- num.cols - length(fixed.col.names)
  result.col.names <- paste('T', 1:num.results, sep='')
  names(csv) <- c(fixed.col.names, result.col.names)
  return(melt(csv, id=c('Site', 'When'), measure=result.col.names))
}


PageCyclerPerms <- function(before, after, f=median, n=10^5-1) {
  # Computes a statistic and P-value for two sets of page_cycler data.
  #
  # Args:
  #   before: A vector of numbers "before" the point of interest.
  #   after: A vector of numbers "after" the point of interest.
  #   f: The statistic to compute. Default is median.
  #   n: The number of permutations to sample. Default is 10^5-1.
  #
  # Returns:
  #   A frame with one row and four columns:
  #   Before: The statistic for the "before" data.
  #   After: The statistic for the "after" data.
  #   Observed: The observed difference.
  #   P.value: The one-sided P-value for the observed difference.

  sample <- numeric(n)

  times <- c(before, after)

  # Find the observed statistic
  stat_before <- f(before)
  stat_after <- f(after)
  observed <- stat_after - stat_before

  for (i in 1:n) {
    index <- sample(length(times), size=length(times) / 2, replace=FALSE)
    sample[i] <- f(times[index]) - f(times[-index])
  }

  p <- (sum(sample >= observed) + 1) / (n + 1)

  # Construct the results
  result <- data.frame()
  result <- rbind(result, c(stat_before, stat_after, observed, p))
  names(result) <- c('Before', 'After', 'Observed', 'P.value')
  return(result)
}
